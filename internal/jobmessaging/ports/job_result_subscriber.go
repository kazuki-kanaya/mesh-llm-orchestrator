package ports

import (
	"context"

	jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
)

type JobResultSubscriber interface {
	Subscribe(ctx context.Context, jobID jobstatedomain.JobID) (JobResultSubscription, error)
}

type JobResultSubscription interface {
	Wait(ctx context.Context) error
	Close() error
}
