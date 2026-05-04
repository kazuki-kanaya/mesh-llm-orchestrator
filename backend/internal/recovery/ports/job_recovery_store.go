package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type JobRecoveryStore interface {
	FindStaleRunning(ctx context.Context, cutoff time.Time, limit int64) ([]uuid.UUID, error)
	RecoverStale(ctx context.Context, jobID uuid.UUID, cutoff time.Time) (bool, error)
}
