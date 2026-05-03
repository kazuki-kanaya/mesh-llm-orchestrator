package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
)

type ExecuteJobUseCase struct {
	repo   ports.JobRepository
	queue  ports.JobQueue
	client ports.HTTPClient
	pub    ports.JobPublisher
}

func NewExecuteJobUseCase(repo ports.JobRepository, queue ports.JobQueue, client ports.HTTPClient, pub ports.JobPublisher) *ExecuteJobUseCase {
	return &ExecuteJobUseCase{
		repo:   repo,
		queue:  queue,
		client: client,
		pub:    pub,
	}
}

func (uc *ExecuteJobUseCase) Execute(ctx context.Context) error {
	jobID, err := uc.queue.Dequeue(ctx)
	if err != nil {
		return err
	}

	claimed, err := uc.repo.Claim(ctx, jobID)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	job, err := uc.repo.Get(ctx, jobID)
	if err != nil {
		return uc.failAndPublish(ctx, jobID)
	}

	resp, err := uc.client.Do(ctx, job.Request)
	if resp == nil || err != nil {
		return uc.failAndPublish(ctx, jobID)
	}

	markedCompleted, err := uc.repo.Complete(ctx, jobID, *resp)
	if err != nil {
		return err
	}
	if !markedCompleted {
		// Another executor may have already moved the job to a terminal state.
		return nil
	}

	return uc.pub.Publish(ctx, jobID)
}

func (uc *ExecuteJobUseCase) failAndPublish(ctx context.Context, jobID uuid.UUID) error {
	markedFailed, err := uc.repo.Fail(ctx, jobID)
	if err != nil {
		return err
	}
	if !markedFailed {
		// Another executor may have already moved the job to a terminal state.
		return nil
	}

	return uc.pub.Publish(ctx, jobID)
}
