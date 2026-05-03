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

type RedisJobRepository struct {
	rdb *goredis.Client
}

func NewRedisJobRepository(rdb *goredis.Client) *RedisJobRepository {
	return &RedisJobRepository{
		rdb: rdb,
	}
}

func (repo *RedisJobRepository) Create(ctx context.Context, job *jobdomain.Job) error {
	values, err := jobinfra.ToRedisHash(job)
	if err != nil {
		return err
	}

	return repo.rdb.HSet(ctx, redis.JobKey(job.ID), values).Err()
}

func (repo *RedisJobRepository) Get(ctx context.Context, jobID uuid.UUID) (*jobdomain.Job, error) {
	values, err := repo.rdb.HGetAll(ctx, redis.JobKey(jobID)).Result()
	if err != nil {
		return nil, err
	}

	return jobinfra.FromRedisHash(jobID, values)
}

var _ ports.JobRepository = (*RedisJobRepository)(nil)
