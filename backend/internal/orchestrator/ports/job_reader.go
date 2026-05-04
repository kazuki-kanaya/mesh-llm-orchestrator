package ports

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type JobReader interface {
	Get(ctx context.Context, jobID uuid.UUID) (*jobdomain.Job, error)
}
