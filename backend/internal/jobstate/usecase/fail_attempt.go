package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type FailAttemptUseCase struct {
	store ports.JobStateStore
}

func NewFailAttemptUseCase(store ports.JobStateStore) *FailAttemptUseCase {
	return &FailAttemptUseCase{
		store: store,
	}
}

type FailAttemptInput struct {
	JobID   domain.JobID
	Attempt int64
}

type FailAttemptOutput struct {
	Accepted bool
}

func (uc *FailAttemptUseCase) Execute(ctx context.Context, input FailAttemptInput) (*FailAttemptOutput, error) {
	if err := input.JobID.Validate(); err != nil {
		return nil, err
	}

	accepted, err := uc.store.FailAttempt(ctx, input.JobID, input.Attempt, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &FailAttemptOutput{
		Accepted: accepted,
	}, nil
}
