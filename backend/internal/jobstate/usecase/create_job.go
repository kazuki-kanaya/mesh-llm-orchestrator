package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type CreateJobUseCase struct {
	repo ports.JobRepository
}

func NewCreateJobUseCase(repo ports.JobRepository) *CreateJobUseCase {
	return &CreateJobUseCase{
		repo: repo,
	}
}

type CreateJobRequest struct {
	Request domain.HTTPRequest
}

type CreateJobResponse struct {
	JobID uuid.UUID
}

func (uc *CreateJobUseCase) Execute(ctx context.Context, request CreateJobRequest) (*CreateJobResponse, error) {
	jobID := uuid.New()
	now := time.Now().UTC()

	job := domain.NewJob(jobID, request.Request, now)
	err := uc.repo.CreateAndEnqueue(ctx, job)
	if err != nil {
		return nil, err
	}

	return &CreateJobResponse{
		JobID: jobID,
	}, nil
}
