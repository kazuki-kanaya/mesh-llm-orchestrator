package ports

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type JobRepository interface {
	Create(ctx context.Context, jobID uuid.UUID, req []byte) error
	Get(ctx context.Context, jobID uuid.UUID) (*jobdomain.Job, error)
}
