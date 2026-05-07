package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobRepository struct {
	rdb *goredis.Client
}

const jobResultPayload = "done"

func NewRedisJobRepository(rdb *goredis.Client) *RedisJobRepository {
	return &RedisJobRepository{
		rdb: rdb,
	}
}

var createAndEnqueueScript = goredis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
	return 0
end

redis.call("HSET", KEYS[1],
	"status", ARGV[1],
	"request", ARGV[2],
	"created_at", ARGV[3],
	"current_attempt", ARGV[4]
)

local stream_id = redis.call("XADD", KEYS[2], "*",
	"job_id", ARGV[5]
)

redis.call("HSET", KEYS[1],
	"stream_id", stream_id
)

return 1
`)

func (r *RedisJobRepository) CreateAndEnqueue(ctx context.Context, job *domain.Job) error {
	values, err := ToRedisHash(job)
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
		values["created_at"],
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

var startAttemptScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") ~= ARGV[1] then
	return {0, 0}
end

local attempt = redis.call("HINCRBY", KEYS[1], "current_attempt", 1)
redis.call("HSET", KEYS[1],
	"status", ARGV[2],
	"started_at", ARGV[3]
)

return {1, attempt}
`)

func (r *RedisJobRepository) StartAttempt(ctx context.Context, jobID domain.JobID, now time.Time) (accepted bool, attempt int64, err error) {
	result, err := startAttemptScript.Run(
		ctx,
		r.rdb,
		[]string{
			redis.JobKey(jobID),
		},
		domain.StatusQueued.String(),
		domain.StatusRunning.String(),
		now.UTC().Format(time.RFC3339Nano),
	).Slice()
	if err != nil {
		return false, 0, err
	}

	if len(result) != 2 {
		return false, 0, fmt.Errorf("invalid start attempt script result: %v", result)
	}

	acceptedInt, ok := result[0].(int64)
	if !ok {
		return false, 0, fmt.Errorf("invalid start attempt accepted result: %v", result[0])
	}

	attempt, ok = result[1].(int64)
	if !ok {
		return false, 0, fmt.Errorf("invalid start attempt value result: %v", result[1])
	}

	return acceptedInt == 1, attempt, nil
}

var completeAttemptScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") ~= ARGV[1] then
	return 0
end

if tonumber(redis.call("HGET", KEYS[1], "current_attempt")) ~= tonumber(ARGV[2]) then
	return 0
end

redis.call("HSET", KEYS[1],
	"response", ARGV[3],
	"status", ARGV[4],
	"terminated_at", ARGV[5]
)

redis.call("PUBLISH", KEYS[2], ARGV[6])

return 1
`)

func (r *RedisJobRepository) CompleteAttempt(ctx context.Context, jobID domain.JobID, attempt int64, response domain.HTTPResponse, now time.Time) (accepted bool, err error) {
	responseBytes, err := json.Marshal(response)
	if err != nil {
		return false, err
	}

	result, err := completeAttemptScript.Run(
		ctx,
		r.rdb,
		[]string{
			redis.JobKey(jobID),
			redis.JobResultChannel(jobID),
		},
		domain.StatusRunning.String(),
		attempt,
		responseBytes,
		domain.StatusCompleted.String(),
		now.UTC().Format(time.RFC3339Nano),
		jobResultPayload,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (r *RedisJobRepository) FailAttempt(ctx context.Context, jobID domain.JobID, attempt int64, now time.Time) (accepted bool, err error) {

}

func (r *RedisJobRepository) Get(ctx context.Context, jobID domain.JobID) (*domain.Job, error) {

}
