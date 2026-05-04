package ports

import (
	"context"

	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type JobCreator interface {
	// Create persists the job and enqueues it atomically.
	// Implementations must not leave a persisted job without a queue entry.
	Create(ctx context.Context, job *jobdomain.Job) error
}
