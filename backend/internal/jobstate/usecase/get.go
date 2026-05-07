package usecase

import (
	"context"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type GetUseCase struct {
	repo ports.JobRepository
}

func NewGetUseCase(repo ports.JobRepository) *GetUseCase {
	return &GetUseCase{
		repo: repo,
	}
}

type GetInput struct {
	JobID domain.JobID
}

type GetOutput struct {
	Job domain.Job
}

func (uc *GetUseCase) Execute(ctx context.Context, input GetInput) (*GetOutput, error) {
	if err := input.JobID.Validate(); err != nil {
		return nil, err
	}

	job, err := uc.repo.Get(ctx, input.JobID)
	if err != nil {
		return nil, err
	}

	return &GetOutput{
		Job: *job,
	}, nil
}
