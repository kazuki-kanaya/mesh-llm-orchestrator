package ports

import (
	"context"

	"github.com/google/uuid"
)

type JobQueue interface {
	Enqueue(ctx context.Context, jobID uuid.UUID) error
}
