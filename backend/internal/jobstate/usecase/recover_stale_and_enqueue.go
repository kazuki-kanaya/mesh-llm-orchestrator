package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type RecoverStaleAndEnqueueUseCase struct {
	store ports.JobStateStore
}

func NewRecoverStaleAndEnqueueUseCase(store ports.JobStateStore) (*RecoverStaleAndEnqueueUseCase, error) {
	if store == nil {
		return nil, ErrNilJobStateStore
	}

	return &RecoverStaleAndEnqueueUseCase{
		store: store,
	}, nil
}

type RecoverStaleAndEnqueueInput struct {
	JobID  domain.JobID
	Cutoff time.Time
}

type RecoverStaleAndEnqueueOutput struct {
	Recovered     bool
	Terminal      bool
	AlreadyQueued bool
}

func (uc *RecoverStaleAndEnqueueUseCase) Execute(ctx context.Context, input RecoverStaleAndEnqueueInput) (*RecoverStaleAndEnqueueOutput, error) {
	if err := input.JobID.Validate(); err != nil {
		return nil, err
	}

	result, err := uc.store.RecoverStaleAndEnqueue(ctx, input.JobID, input.Cutoff.UTC())
	if err != nil {
		return nil, err
	}

	return &RecoverStaleAndEnqueueOutput{
		Recovered:     result.Recovered,
		Terminal:      result.Terminal,
		AlreadyQueued: result.AlreadyQueued,
	}, nil
}
