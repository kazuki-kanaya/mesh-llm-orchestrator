package usecase

import (
	"context"
	"time"

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

type CreateJobInput struct {
	Request domain.HTTPRequest
}

type CreateJobOutput struct {
	JobID domain.JobID
}

func (uc *CreateJobUseCase) Execute(ctx context.Context, input CreateJobInput) (*CreateJobOutput, error) {
	jobID := domain.NewJobID()

	job, err := domain.NewJob(jobID, input.Request, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	err = uc.repo.CreateAndEnqueue(ctx, job)
	if err != nil {
		return nil, err
	}

	return &CreateJobOutput{
		JobID: jobID,
	}, nil
}
