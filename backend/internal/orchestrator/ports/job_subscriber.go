package ports

import (
	"context"

	"github.com/google/uuid"
)

type JobSubscriber interface {
	Subscribe(ctx context.Context, jobID uuid.UUID) (Subscription, error)
}

type Subscription interface {
	Channel() <-chan struct{}
	Close() error
}
