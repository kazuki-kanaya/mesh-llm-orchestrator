package ports

import (
	"context"
	"time"

	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

type JobReconcileClient interface {
	RecoverStaleAndEnqueue(ctx context.Context, jobID jobstatedomain.JobID, cutoff time.Time) (jobstatedomain.StaleJobRecoveryResult, error)
}
