package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/ports"
)

type CreateJobUseCase struct {
	repo ports.JobRepository
}

type CreateJobInput struct {
	Model            string
	Messages         json.RawMessage
	GenerationParams json.RawMessage
}

type CreateJobOutput struct {
	JobID  uuid.UUID
	Status domain.Status
}

func NewCreateJobUseCase(repo ports.JobRepository) *CreateJobUseCase {
	return &CreateJobUseCase{
		repo: repo,
	}
}

func (uc *CreateJobUseCase) Execute(ctx context.Context, input CreateJobInput) (*CreateJobOutput, error) {
	now := time.Now()
	jobID := uuid.New()

	job := domain.NewJob(
		jobID,
		input.Model,
		input.Messages,
		input.GenerationParams,
		now,
	)

	if err := uc.repo.Create(ctx, job); err != nil {
		return nil, err
	}

	return &CreateJobOutput{
		JobID:  job.JobID,
		Status: job.Status,
	}, nil
}
