package ports

import (
	"context"

	"github.com/google/uuid"
)

type JobPublisher interface {
	Publish(ctx context.Context, jobID uuid.UUID) error
}
