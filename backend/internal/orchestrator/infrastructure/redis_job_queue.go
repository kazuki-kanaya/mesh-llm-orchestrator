package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobQueue struct {
	rdb *goredis.Client
}

func NewRedisJobQueue(rdb *goredis.Client) *RedisJobQueue {
	return &RedisJobQueue{
		rdb: rdb,
	}
}

func (q *RedisJobQueue) Enqueue(ctx context.Context, jobID uuid.UUID) error {
	return q.rdb.RPush(ctx, redis.JobQueueKey(), jobID.String()).Err()
}

var _ ports.JobQueue = (*RedisJobQueue)(nil)
