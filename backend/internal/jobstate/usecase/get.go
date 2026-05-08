package usecase

import (
	"context"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type GetUseCase struct {
	store ports.JobStateStore
}

func NewGetUseCase(store ports.JobStateStore) *GetUseCase {
	return &GetUseCase{
		store: store,
	}
}

type GetInput struct {
	JobID domain.JobID
}

type GetOutput struct {
	Job *domain.Job
}

func (uc *GetUseCase) Execute(ctx context.Context, input GetInput) (*GetOutput, error) {
	if err := input.JobID.Validate(); err != nil {
		return nil, err
	}

	job, err := uc.store.Get(ctx, input.JobID)
	if err != nil {
		return nil, err
	}

	return &GetOutput{
		Job: job,
	}, nil
}
