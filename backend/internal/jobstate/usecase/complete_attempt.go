package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type CompleteAttemptUseCase struct {
	store ports.JobStateStore
}

func NewCompleteAttemptUseCase(store ports.JobStateStore) *CompleteAttemptUseCase {
	return &CompleteAttemptUseCase{
		store: store,
	}
}

type CompleteAttemptInput struct {
	JobID    domain.JobID
	Attempt  int64
	Response *domain.HTTPResponse
}

type CompleteAttemptOutput struct {
	Accepted bool
}

func (uc *CompleteAttemptUseCase) Execute(ctx context.Context, input CompleteAttemptInput) (*CompleteAttemptOutput, error) {
	if err := input.JobID.Validate(); err != nil {
		return nil, err
	}
	if input.Response == nil {
		return nil, domain.ErrNilHTTPResponse
	}

	accepted, err := uc.store.CompleteAttempt(ctx, input.JobID, input.Attempt, input.Response, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &CompleteAttemptOutput{
		Accepted: accepted,
	}, nil
}
