package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobPublisher struct {
	rdb *goredis.Client
}

const jobResultPayload = "done"

func NewRedisJobPublisher(rdb *goredis.Client) ports.JobPublisher {
	return &RedisJobPublisher{
		rdb: rdb,
	}
}

func (p *RedisJobPublisher) Publish(ctx context.Context, jobID uuid.UUID) error {
	return p.rdb.Publish(ctx, redis.JobResultChannel(jobID), jobResultPayload).Err()
}

var _ ports.JobPublisher = (*RedisJobPublisher)(nil)
