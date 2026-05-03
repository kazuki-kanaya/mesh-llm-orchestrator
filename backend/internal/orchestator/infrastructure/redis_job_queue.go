package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisJobQueue struct {
	rdb *redis.Client
}

func NewRedisJobQueue(rdb *redis.Client) *RedisJobQueue {
	return &RedisJobQueue{
		rdb: rdb,
	}
}

func (q *RedisJobQueue) Enqueue(ctx context.Context, jobID uuid.UUID) error {
	return q.rdb.RPush(ctx, jobQueueKey(), jobID.String()).Err()
}

func jobQueueKey() string {
	return "queue:jobs"
}
