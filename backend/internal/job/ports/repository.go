package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type JobRepository interface {
	Create(ctx context.Context, job *domain.Job) error
	GetByID(ctx context.Context, jobID uuid.UUID) (*domain.Job, error)
	Update(ctx context.Context, job *domain.Job) error
}
