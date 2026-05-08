package ports

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/domain"
)

type JobQueue interface {
	EnsureGroup(ctx context.Context) error
	Read(ctx context.Context, consumerName domain.ConsumerName) (*domain.Message, error)
	ClaimStalePending(ctx context.Context, consumerName domain.ConsumerName, minIdle time.Duration, count int64) ([]*domain.Message, error)
	Ack(ctx context.Context, messageID domain.MessageID) error
}
