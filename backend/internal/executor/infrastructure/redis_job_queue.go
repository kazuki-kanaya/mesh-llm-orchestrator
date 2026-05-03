package infrastructure

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobQueue struct {
	rdb *goredis.Client
}

func NewRedisJobQueue(rdb *goredis.Client) ports.JobQueue {
	return &RedisJobQueue{
		rdb: rdb,
	}
}

func (q *RedisJobQueue) Dequeue(ctx context.Context) (uuid.UUID, error) {
	result, err := q.rdb.BLPop(ctx, 0, redis.JobQueueKey()).Result()
	if err != nil {
		return uuid.Nil, err
	}

	if len(result) < 2 {
		return uuid.Nil, errors.New("invalid queue response")
	}

	jobID, err := uuid.Parse(result[1])
	if err != nil {
		return uuid.Nil, err
	}

	return jobID, nil
}

var _ ports.JobQueue = (*RedisJobQueue)(nil)
