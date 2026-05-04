package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/recovery/ports"
)

type RecoverStaleJobsUseCase struct {
	store      ports.JobRecoveryStore
	staleAfter time.Duration
	limit      int64
}

func NewRecoverStaleJobsUseCase(store ports.JobRecoveryStore, staleAfter time.Duration, limit int64) *RecoverStaleJobsUseCase {
	return &RecoverStaleJobsUseCase{
		store:      store,
		staleAfter: staleAfter,
		limit:      limit,
	}
}

func (uc *RecoverStaleJobsUseCase) Execute(ctx context.Context, now time.Time) (int, error) {
	cutoff := now.Add(-uc.staleAfter)

	jobIDs, err := uc.store.FindStaleRunning(ctx, cutoff, uc.limit)
	if err != nil {
		return 0, err
	}

	recovered := 0
	for _, jobID := range jobIDs {
		ok, err := uc.store.RecoverStale(ctx, jobID, cutoff)
		if err != nil {
			return recovered, err
		}
		if ok {
			recovered++
		}
	}

	return recovered, nil
}
