package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
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

var createAndEnqueueScript = goredis.NewScript(`
redis.call("HSET", KEYS[1],
	"status", ARGV[1],
	"request", ARGV[2],
	"current_attempt", ARGV[3]
)

local stream_id = redis.call("XADD", KEYS[2], "*",
	"job_id", ARGV[4]
)

redis.call("HSET", KEYS[1],
	"stream_id", stream_id
)

return 1
`)

func (r *RedisJobRepository) CreateAndEnqueue(ctx context.Context, job domain.Job) error {
	values, err := ToRedisHash(&job)
	if err != nil {
		return err
	}
	result, err := createAndEnqueueScript.Run(
		ctx,
		r.rdb,
		[]string{
			redis.JobKey(job.ID),
			redis.JobStreamKey(),
		},
		values["status"],
		values["request"],
		values["current_attempt"],
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

func (r *RedisJobRepository) StartAttempt(ctx context.Context, jobID domain.JobID, now time.Time) (accepted bool, attempt int64, err error) {

}

func (r *RedisJobRepository) CompleteAttempt(ctx context.Context, jobID domain.JobID, attempt int64, response domain.HTTPResponse, now time.Time) (accepted bool, err error) {

}

func (r *RedisJobRepository) FailAttempt(ctx context.Context, jobID domain.JobID, attempt int64, now time.Time) (accepted bool, err error) {

}

func (r *RedisJobRepository) Get(ctx context.Context, jobID domain.JobID) (*domain.Job, error) {

}
