package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type StartAttemptUseCase struct {
	store ports.JobStateStore
}

func NewStartAttemptUseCase(store ports.JobStateStore) (*StartAttemptUseCase, error) {
	if store == nil {
		return nil, ErrNilJobStateStore
	}

	return &StartAttemptUseCase{
		store: store,
	}, nil
}

type StartAttemptInput struct {
	JobID domain.JobID
}

type StartAttemptOutput struct {
	Accepted bool
	Attempt  int64
}

func (uc *StartAttemptUseCase) Execute(ctx context.Context, input StartAttemptInput) (*StartAttemptOutput, error) {
	if err := input.JobID.Validate(); err != nil {
		return nil, err
	}

	accepted, attempt, err := uc.store.StartAttempt(ctx, input.JobID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &StartAttemptOutput{
		Accepted: accepted,
		Attempt:  attempt,
	}, nil
}
