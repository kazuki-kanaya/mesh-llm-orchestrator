package usecase

import (
	"context"

	jobqueuedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/domain"
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

	accepted, attempt, err := s.jobState.ClaimAttempt(ctx, msg.JobID)
	if err != nil {
		return err
	}
	if !accepted {
		job, err := s.jobState.Get(ctx, msg.JobID)
		if err != nil {
			return err
		}
		if job.Status.IsTerminal() {
			if err := s.queue.Ack(ctx, msg.ID); err != nil {
				return err
			}
			return nil
		}
		return nil
	}

	job, err := s.jobState.Get(ctx, msg.JobID)
	if err != nil {
		if _, failErr := s.jobState.FailAttempt(ctx, msg.JobID, attempt); failErr != nil {
			return failErr
		}
		if ackErr := s.queue.Ack(ctx, msg.ID); ackErr != nil {
			return ackErr
		}
		return err
	}

	resp, err := s.executeRequest(ctx, job.Request)
	if err != nil {
		accepted, failErr := s.jobState.FailAttempt(ctx, msg.JobID, attempt)
		if failErr != nil {
			return failErr
		}
		if !accepted {
			return nil
		}

		if ackErr := s.queue.Ack(ctx, msg.ID); ackErr != nil {
			return ackErr
		}

		return err
	}

	if _, err := s.jobState.CompleteAttempt(ctx, msg.JobID, attempt, *resp); err != nil {
		return err
	}

	if err := s.queue.Ack(ctx, msg.ID); err != nil {
		return err
	}

	return nil
}
