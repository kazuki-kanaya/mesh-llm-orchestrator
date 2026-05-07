package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type CompleteAttemptUseCase struct {
	repo ports.JobRepository
}

func NewCompleteAttemptUseCase(repo ports.JobRepository) *CompleteAttemptUseCase {
	return &CompleteAttemptUseCase{
		repo: repo,
	}
}

type CompleteAttemptInput struct {
	JobID    domain.JobID
	Attempt  int64
	Response domain.HTTPResponse
}

type CompleteAttemptOutput struct {
	Accepted bool
}

func (uc *CompleteAttemptUseCase) Execute(ctx context.Context, input CompleteAttemptInput) (*CompleteAttemptOutput, error) {
	if err := input.JobID.Validate(); err != nil {
		return nil, err
	}

	accepted, err := uc.repo.CompleteAttempt(ctx, input.JobID, input.Attempt, input.Response, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &CompleteAttemptOutput{
		Accepted: accepted,
	}, nil
}
