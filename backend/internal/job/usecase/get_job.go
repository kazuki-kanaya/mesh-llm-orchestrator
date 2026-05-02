package usecase

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type GetJobRepository interface {
	GetByID(ctx context.Context, jobID uuid.UUID) (*domain.Job, error)
}

type GetJobUseCase struct {
	repo GetJobRepository
}

type GetJobInput struct {
	JobID uuid.UUID
}

type GetJobOutput struct {
	JobID            uuid.UUID
	Model            string
	GenerationParams json.RawMessage
	Status           domain.Status
	FinalResult      *string
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewGetJobUseCase(repo GetJobRepository) *GetJobUseCase {
	return &GetJobUseCase{
		repo: repo,
	}
}

func (uc *GetJobUseCase) Execute(ctx context.Context, input GetJobInput) (*GetJobOutput, error) {
	job, err := uc.repo.GetByID(ctx, input.JobID)
	if err != nil {
		return nil, err
	}
	return &GetJobOutput{
		JobID:            job.JobID,
		Model:            job.Model,
		GenerationParams: job.GenerationParams,
		Status:           job.Status,
		FinalResult:      job.FinalResult,
		ErrorMessage:     job.ErrorMessage,
		CreatedAt:        job.CreatedAt,
		UpdatedAt:        job.UpdatedAt,
	}, nil
}
