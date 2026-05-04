package infrastructure

import (
	"context"
	"fmt"

	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	jobinfra "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/infrastructure"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobCreator struct {
	rdb *goredis.Client
}

func NewRedisJobCreator(rdb *goredis.Client) ports.JobCreator {
	return &RedisJobCreator{
		rdb: rdb,
	}
}

var createJobScript = goredis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
	return 0
end

redis.call("HSET", KEYS[1],
	"status", ARGV[1],
	"request", ARGV[2],
	"retry_count", ARGV[3]
)

redis.call("RPUSH", KEYS[2], ARGV[4])

return 1
`)

func (c *RedisJobCreator) Create(ctx context.Context, job *jobdomain.Job) error {
	values, err := jobinfra.ToRedisHash(job)
	if err != nil {
		return err
	}

	result, err := createJobScript.Run(
		ctx,
		c.rdb,
		[]string{
			redis.JobKey(job.ID),
			redis.JobQueueKey(),
		},
		values["status"],
		values["request"],
		values["retry_count"],
		job.ID.String(),
	).Int()
	if err != nil {
		return err
	}

	if result == 0 {
		return fmt.Errorf("job already exists: %s", job.ID)
	}

	return nil
}

var _ ports.JobCreator = (*RedisJobCreator)(nil)
