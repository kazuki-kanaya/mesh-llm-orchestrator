package usecase

import (
	"context"

	jobqueuedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/domain"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

type ProcessMessageInput struct {
	ConsumerName jobqueuedomain.ConsumerName
}

func (s *Service) ProcessMessage(ctx context.Context, input ProcessMessageInput) error {
	if err := input.ConsumerName.Validate(); err != nil {
		return err
	}

	msg, err := s.queue.Read(ctx, input.ConsumerName)
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	return s.executeMessage(ctx, msg)
}

// A message is acknowledged only after this executor settles the claimed attempt,
// or when the job is already terminal. Non-terminal unclaimed messages are left
// pending so a reconciler can recover them later.
func (s *Service) executeMessage(ctx context.Context, msg *jobqueuedomain.Message) error {
	accepted, attempt, err := s.jobExecutionClient.ClaimAttempt(ctx, msg.JobID)
	if err != nil {
		return err
	}
	if !accepted {
		return s.ackIfTerminal(ctx, msg)
	}

	job, err := s.jobExecutionClient.Get(ctx, msg.JobID)
	if err != nil {
		// Once an attempt is claimed, failing to load its request is treated as
		// an internal execution failure rather than leaving it for recovery.
		return s.failAndAck(ctx, msg, attempt)
	}

	resp, err := s.executeRequest(ctx, job.Request)
	if err != nil {
		return s.failAndAck(ctx, msg, attempt)
	}

	return s.completeAndAck(ctx, msg, attempt, resp)
}

func (s *Service) ackIfTerminal(ctx context.Context, msg *jobqueuedomain.Message) error {
	job, err := s.jobExecutionClient.Get(ctx, msg.JobID)
	if err != nil {
		return err
	}

	if !job.Status.IsTerminal() {
		return nil
	}

	return s.queue.Ack(ctx, msg.ID)
}

func (s *Service) failAndAck(ctx context.Context, msg *jobqueuedomain.Message, attempt int64) error {
	accepted, err := s.jobExecutionClient.FailAttempt(ctx, msg.JobID, attempt)
	if err != nil {
		return err
	}
	if !accepted {
		return s.ackIfTerminal(ctx, msg)
	}

	if err := s.queue.Ack(ctx, msg.ID); err != nil {
		return err
	}

	return nil
}

func (s *Service) completeAndAck(
	ctx context.Context,
	msg *jobqueuedomain.Message,
	attempt int64,
	resp *jobstatedomain.HTTPResponse,
) error {
	accepted, err := s.jobExecutionClient.CompleteAttempt(ctx, msg.JobID, attempt, resp)
	if err != nil {
		return err
	}
	if !accepted {
		return s.ackIfTerminal(ctx, msg)
	}

	return s.queue.Ack(ctx, msg.ID)
}
