package usecase

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/ports"
)

type CreateJobUseCase struct {
	creator ports.JobCreator
}

func NewCreateJobUseCase(creator ports.JobCreator) *CreateJobUseCase {
	return &CreateJobUseCase{
		creator: creator,
	}
}

func (uc *CreateJobUseCase) Execute(ctx context.Context, req jobdomain.HTTPRequest) (uuid.UUID, error) {
	jobID := uuid.New()
	job := jobdomain.NewJob(jobID, req)

	if err := uc.creator.Create(ctx, job); err != nil {
		return uuid.Nil, err
	}

	return jobID, nil
}
