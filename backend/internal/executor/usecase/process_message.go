package usecase

import (
	"context"

	jobqueuedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/domain"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

type ProcessMessageInput struct {
	ConsumerName jobqueuedomain.ConsumerName
}

type ProcessMessageOutput struct {
	Processed bool
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

	return s.processMessage(ctx, msg)
}

// A message is acknowledged only after this executor settles the claimed attempt,
// or when the job is already terminal. Non-terminal unclaimed messages are left
// pending so a reconciler can recover them later.
func (s *Service) processMessage(ctx context.Context, msg *jobqueuedomain.Message) error {
	accepted, attempt, err := s.jobState.ClaimAttempt(ctx, msg.JobID)
	if err != nil {
		return err
	}
	if !accepted {
		return s.ackIfTerminal(ctx, msg)
	}

	job, err := s.jobState.Get(ctx, msg.JobID)
	if err != nil {
		return s.failAndAck(ctx, msg, attempt, err)
	}

	resp, err := s.executeRequest(ctx, job.Request)
	if err != nil {
		return s.failAndAck(ctx, msg, attempt, err)
	}

	return s.completeAndAck(ctx, msg, attempt, resp)
}

func (s *Service) ackIfTerminal(ctx context.Context, msg *jobqueuedomain.Message) error {
	job, err := s.jobState.Get(ctx, msg.JobID)
	if err != nil {
		return err
	}

	if !job.Status.IsTerminal() {
		return nil
	}

	return s.queue.Ack(ctx, msg.ID)
}

func (s *Service) failAndAck(ctx context.Context, msg *jobqueuedomain.Message, attempt int64, cause error) error {
	accepted, err := s.jobState.FailAttempt(ctx, msg.JobID, attempt)
	if err != nil {
		return err
	}
	if !accepted {
		return s.ackIfTerminal(ctx, msg)
	}

	if err := s.queue.Ack(ctx, msg.ID); err != nil {
		return err
	}

	return cause
}

func (s *Service) completeAndAck(
	ctx context.Context,
	msg *jobqueuedomain.Message,
	attempt int64,
	resp *jobstatedomain.HTTPResponse,
) error {
	accepted, err := s.jobState.CompleteAttempt(ctx, msg.JobID, attempt, resp)
	if err != nil {
		return err
	}
	if !accepted {
		return s.ackIfTerminal(ctx, msg)
	}

	return s.queue.Ack(ctx, msg.ID)
}
