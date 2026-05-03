package ports

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type JobRepository interface {
	Claim(ctx context.Context, jobID uuid.UUID) (bool, error)
	Get(ctx context.Context, jobID uuid.UUID) (*jobdomain.Job, error)
	Complete(ctx context.Context, jobID uuid.UUID, resp jobdomain.HTTPResponse) (bool, error)
	Fail(ctx context.Context, jobID uuid.UUID) (bool, error)
}
