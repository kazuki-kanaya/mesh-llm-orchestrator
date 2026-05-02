package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/ports"
)

type CancelJobUseCase struct {
	repo ports.JobRepository
}

type CancelJobInput struct {
	JobID uuid.UUID
}

type CancelJobOutput struct {
	JobID  uuid.UUID
	Status domain.Status
}

func NewCancelJobUseCase(repo ports.JobRepository) *CancelJobUseCase {
	return &CancelJobUseCase{
		repo: repo,
	}
}

func (uc *CancelJobUseCase) Execute(ctx context.Context, input CancelJobInput) (*CancelJobOutput, error) {
	job, err := uc.repo.GetByID(ctx, input.JobID)
	if err != nil {
		return nil, err
	}

	if !job.Status.IsCancelable() {
		return &CancelJobOutput{
			JobID:  job.JobID,
			Status: job.Status,
		}, nil
	}

	job.Status = domain.StatusCancelled
	job.UpdatedAt = time.Now()

	if err := uc.repo.Update(ctx, job); err != nil {
		return nil, err
	}

	return &CancelJobOutput{
		JobID:  job.JobID,
		Status: job.Status,
	}, nil
}
