package ports

import (
	"context"

	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

type JobRequestClient interface {
	CreateAndEnqueue(ctx context.Context, job *jobstatedomain.Job) error
	Wait(ctx context.Context, jobID jobstatedomain.JobID) (*jobstatedomain.Job, error)
}
