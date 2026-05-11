package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/kazuki-kanaya/quegress/internal/executor/domain"
	"github.com/kazuki-kanaya/quegress/internal/executor/ports"
	jobmessagingdomain "github.com/kazuki-kanaya/quegress/internal/jobmessaging/domain"
	jobmessagingports "github.com/kazuki-kanaya/quegress/internal/jobmessaging/ports"
	jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
)

var (
	ErrNilJobQueue           = errors.New("job queue is nil")
	ErrNilJobExecutionClient = errors.New("job execution client is nil")
	ErrNilHTTPClient         = errors.New("http client is nil")
	ErrInvalidRequestTimeout = errors.New("request timeout must be positive")
)

type ProcessMessageUseCase struct {
	queue              jobmessagingports.JobQueue
	jobExecutionClient ports.JobExecutionClient
	httpClient         ports.HTTPClient
	requestTimeout     time.Duration
}

func NewProcessMessageUseCase(
	queue jobmessagingports.JobQueue,
	jobExecutionClient ports.JobExecutionClient,
	httpClient ports.HTTPClient,
	requestTimeout time.Duration,
) (*ProcessMessageUseCase, error) {
	if queue == nil {
		return nil, ErrNilJobQueue
	}
	if jobExecutionClient == nil {
		return nil, ErrNilJobExecutionClient
	}
	if httpClient == nil {
		return nil, ErrNilHTTPClient
	}
	if requestTimeout <= 0 {
		return nil, ErrInvalidRequestTimeout
	}

	return &ProcessMessageUseCase{
		queue:              queue,
		jobExecutionClient: jobExecutionClient,
		httpClient:         httpClient,
		requestTimeout:     requestTimeout,
	}, nil
}

type ProcessMessageInput struct {
	ConsumerName jobmessagingdomain.ConsumerName
}

func (uc *ProcessMessageUseCase) Execute(ctx context.Context, input ProcessMessageInput) error {
	if err := input.ConsumerName.Validate(); err != nil {
		return err
	}

	msg, err := uc.queue.Read(ctx, input.ConsumerName)
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	// A message is acknowledged only after this executor settles the claimed attempt,
	// or when the job is already terminal. Non-terminal unclaimed messages are left
	// pending so a reconciler can recover them later.
	accepted, attempt, err := uc.jobExecutionClient.ClaimAttempt(ctx, msg.JobID)
	if err != nil {
		return err
	}
	if !accepted {
		return uc.ackIfTerminal(ctx, msg)
	}

	job, err := uc.jobExecutionClient.Get(ctx, msg.JobID)
	if err != nil {
		// Once an attempt is claimed, failing to load its request is treated as
		// an internal execution failure rather than leaving it for recovery.
		return uc.failAndAck(ctx, msg, attempt)
	}

	resp, err := uc.executeRequest(ctx, job.Request)
	if err != nil {
		return uc.failAndAck(ctx, msg, attempt)
	}

	return uc.completeAndAck(ctx, msg, attempt, resp)
}

func (uc *ProcessMessageUseCase) ackIfTerminal(ctx context.Context, msg *jobmessagingdomain.Message) error {
	job, err := uc.jobExecutionClient.Get(ctx, msg.JobID)
	if err != nil {
		return err
	}

	if !job.Status.IsTerminal() {
		return nil
	}

	return uc.queue.Ack(ctx, msg.ID)
}

func (uc *ProcessMessageUseCase) failAndAck(ctx context.Context, msg *jobmessagingdomain.Message, attempt int64) error {
	accepted, err := uc.jobExecutionClient.FailAttempt(ctx, msg.JobID, attempt)
	if err != nil {
		return err
	}
	if !accepted {
		return uc.ackIfTerminal(ctx, msg)
	}

	return uc.queue.Ack(ctx, msg.ID)
}

func (uc *ProcessMessageUseCase) completeAndAck(
	ctx context.Context,
	msg *jobmessagingdomain.Message,
	attempt int64,
	resp *jobstatedomain.HTTPResponse,
) error {
	accepted, err := uc.jobExecutionClient.CompleteAttempt(ctx, msg.JobID, attempt, resp)
	if err != nil {
		return err
	}
	if !accepted {
		return uc.ackIfTerminal(ctx, msg)
	}

	return uc.queue.Ack(ctx, msg.ID)
}

func (uc *ProcessMessageUseCase) executeRequest(ctx context.Context, req jobstatedomain.HTTPRequest) (*jobstatedomain.HTTPResponse, error) {
	requestCtx, cancel := context.WithTimeout(ctx, uc.requestTimeout)
	defer cancel()

	resp, err := uc.httpClient.Do(requestCtx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, domain.ErrNilHTTPResponse
	}
	return resp, nil
}
