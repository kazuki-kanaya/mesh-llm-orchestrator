package ports

import (
	"context"

	"github.com/google/uuid"
)

type JobRepository interface {
	Create(ctx context.Context, jobID uuid.UUID, req []byte) error
}
