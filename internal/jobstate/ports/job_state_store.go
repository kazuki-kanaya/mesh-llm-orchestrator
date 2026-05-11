package ports

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
)

type JobStateStore interface {
	CreateAndEnqueue(ctx context.Context, job *domain.Job) error
	StartAttempt(ctx context.Context, jobID domain.JobID, now time.Time) (accepted bool, attempt int64, err error)

	// CompleteAttempt marks the attempt completed, stores the response, and notifies waiters.
	CompleteAttempt(ctx context.Context, jobID domain.JobID, attempt int64, response *domain.HTTPResponse, now time.Time) (accepted bool, err error)

	// FailAttempt marks the attempt failed and notifies waiters.
	FailAttempt(ctx context.Context, jobID domain.JobID, attempt int64, now time.Time) (accepted bool, err error)

	RecoverStaleAndEnqueue(ctx context.Context, jobID domain.JobID, cutoff time.Time) (domain.StaleJobRecoveryResult, error)
	Get(ctx context.Context, jobID domain.JobID) (*domain.Job, error)
}
