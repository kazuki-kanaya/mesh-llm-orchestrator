package usecase

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestator/ports"
)

type CreateJobUseCase struct {
	repo  ports.JobRepository
	queue ports.JobQueue
}

func NewCreateJobUseCase(repo ports.JobRepository, queue ports.JobQueue) *CreateJobUseCase {
	return &CreateJobUseCase{
		repo:  repo,
		queue: queue,
	}
}

func (uc *CreateJobUseCase) Execute(ctx context.Context, req jobdomain.HTTPRequest) (uuid.UUID, error) {
	jobID := uuid.New()
	job := &jobdomain.Job{
		ID:         jobID,
		Status:     jobdomain.StatusQueued,
		Request:    req,
		RetryCount: 0,
	}

	if err := uc.repo.Create(ctx, job); err != nil {
		return uuid.Nil, err
	}

	if err := uc.queue.Enqueue(ctx, jobID); err != nil {
		return uuid.Nil, err
	}

	return jobID, nil
}
