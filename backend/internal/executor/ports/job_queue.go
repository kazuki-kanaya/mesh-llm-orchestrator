package ports

import (
	"context"

	"github.com/google/uuid"
)

type JobQueue interface {
	Dequeue(ctx context.Context) (uuid.UUID, error)
}
