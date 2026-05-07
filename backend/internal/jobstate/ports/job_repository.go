package ports

import (
	"context"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

type JobRepository interface {
	CreateAndEnqueue(ctx context.Context, job *domain.Job) error
}
