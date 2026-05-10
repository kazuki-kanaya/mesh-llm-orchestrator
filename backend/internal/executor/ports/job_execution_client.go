package ports

import (
	"context"

	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

type JobExecutionClient interface {
	ClaimAttempt(ctx context.Context, jobID jobstatedomain.JobID) (accepted bool, attempt int64, err error)
	Get(ctx context.Context, jobID jobstatedomain.JobID) (*jobstatedomain.Job, error)
	CompleteAttempt(ctx context.Context, jobID jobstatedomain.JobID, attempt int64, response *jobstatedomain.HTTPResponse) (accepted bool, err error)
	FailAttempt(ctx context.Context, jobID jobstatedomain.JobID, attempt int64) (accepted bool, err error)
}
