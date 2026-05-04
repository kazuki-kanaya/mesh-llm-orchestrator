package infrastructure

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	jobinfra "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/infrastructure"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobReader struct {
	rdb *goredis.Client
}

func NewRedisJobReader(rdb *goredis.Client) ports.JobReader {
	return &RedisJobReader{
		rdb: rdb,
	}
}

func (r *RedisJobReader) Get(ctx context.Context, jobID uuid.UUID) (*jobdomain.Job, error) {
	values, err := r.rdb.HGetAll(ctx, redis.JobKey(jobID)).Result()
	if err != nil {
		return nil, err
	}

	return jobinfra.FromRedisHash(jobID, values)
}

var _ ports.JobReader = (*RedisJobReader)(nil)
