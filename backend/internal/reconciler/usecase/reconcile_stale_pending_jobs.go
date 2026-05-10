package usecase

import (
	"context"
	"errors"
	"time"

	jobmessagingdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/domain"
	jobmessagingports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/reconciler/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/reconciler/ports"
)

var (
	ErrNilJobQueue           = errors.New("job queue is nil")
	ErrNilJobReconcileClient = errors.New("job reconcile client is nil")
)

type ReconcileStalePendingJobsUseCase struct {
	queue      jobmessagingports.JobQueue
	client     ports.JobReconcileClient
	staleAfter time.Duration
	batchSize  int64
}

func NewReconcileStalePendingJobsUseCase(
	queue jobmessagingports.JobQueue,
	client ports.JobReconcileClient,
	staleAfter time.Duration,
	batchSize int64,
) (*ReconcileStalePendingJobsUseCase, error) {
	if queue == nil {
		return nil, ErrNilJobQueue
	}
	if client == nil {
		return nil, ErrNilJobReconcileClient
	}
	if staleAfter <= 0 {
		return nil, domain.ErrInvalidStaleAfter
	}
	if batchSize <= 0 {
		return nil, domain.ErrInvalidBatchSize
	}

	return &ReconcileStalePendingJobsUseCase{
		queue:      queue,
		client:     client,
		staleAfter: staleAfter,
		batchSize:  batchSize,
	}, nil
}

type ReconcileStalePendingJobsInput struct {
	ConsumerName jobmessagingdomain.ConsumerName
}

type ReconcileStalePendingJobsOutput struct {
	Claimed       int
	Recovered     int
	AckedTerminal int
	AckedQueued   int
}

func (uc *ReconcileStalePendingJobsUseCase) Execute(ctx context.Context, input ReconcileStalePendingJobsInput) (*ReconcileStalePendingJobsOutput, error) {
	if err := input.ConsumerName.Validate(); err != nil {
		return nil, err
	}

	messages, err := uc.queue.ClaimStalePending(ctx, input.ConsumerName, uc.staleAfter, uc.batchSize)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().UTC().Add(-uc.staleAfter)
	output := &ReconcileStalePendingJobsOutput{
		Claimed: len(messages),
	}

	for _, msg := range messages {
		result, err := uc.client.RecoverStaleAndEnqueue(ctx, msg.JobID, cutoff)
		if err != nil {
			return nil, err
		}

		switch {
		case result.Recovered:
			if err := uc.queue.Ack(ctx, msg.ID); err != nil {
				return nil, err
			}
			output.Recovered++
		case result.Terminal:
			if err := uc.queue.Ack(ctx, msg.ID); err != nil {
				return nil, err
			}
			output.AckedTerminal++
		case result.AlreadyQueued:
			if err := uc.queue.Ack(ctx, msg.ID); err != nil {
				return nil, err
			}
			output.AckedQueued++
		}
	}

	return output, nil
}
