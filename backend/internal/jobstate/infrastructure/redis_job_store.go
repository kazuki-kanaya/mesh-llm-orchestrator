package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
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

var ErrNilRedisClient = errors.New("redis client is nil")

func NewRedisJobStore(rdb *goredis.Client) (*RedisJobRepository, error) {
	if rdb == nil {
		return nil, ErrNilRedisClient
	}

	return &RedisJobRepository{
		rdb: rdb,
	}, nil
}

// Keep this field list in sync with ToRedisHash. Creation is scripted so the
// job hash and initial stream message are written atomically.
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

redis.call("XADD", KEYS[2], "*",
	"job_id", ARGV[5]
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
			redis.JobKey(job.ID.String()),
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
	"started_at", ARGV[3],
	"started_at_unix_milli", ARGV[4]
)

return {1, attempt}
`)

func (r *RedisJobRepository) StartAttempt(ctx context.Context, jobID domain.JobID, now time.Time) (accepted bool, attempt int64, err error) {
	result, err := startAttemptScript.Run(
		ctx,
		r.rdb,
		[]string{
			redis.JobKey(jobID.String()),
		},
		domain.StatusQueued.String(),
		domain.StatusRunning.String(),
		now.UTC().Format(time.RFC3339Nano),
		now.UTC().UnixMilli(),
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

func (r *RedisJobRepository) CompleteAttempt(ctx context.Context, jobID domain.JobID, attempt int64, response *domain.HTTPResponse, now time.Time) (accepted bool, err error) {
	if response == nil {
		return false, domain.ErrNilHTTPResponse
	}

	responseBytes, err := json.Marshal(response)
	if err != nil {
		return false, err
	}

	result, err := completeAttemptScript.Run(
		ctx,
		r.rdb,
		[]string{
			redis.JobKey(jobID.String()),
			redis.JobResultChannel(jobID.String()),
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

var failAttemptScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") ~= ARGV[1] then
	return 0
end

if tonumber(redis.call("HGET", KEYS[1], "current_attempt")) ~= tonumber(ARGV[2]) then
	return 0
end

redis.call("HSET", KEYS[1],
	"status", ARGV[3],
	"terminated_at", ARGV[4]
)

redis.call("PUBLISH", KEYS[2], ARGV[5])

return 1
`)

func (r *RedisJobRepository) FailAttempt(ctx context.Context, jobID domain.JobID, attempt int64, now time.Time) (accepted bool, err error) {
	result, err := failAttemptScript.Run(
		ctx,
		r.rdb,
		[]string{
			redis.JobKey(jobID.String()),
			redis.JobResultChannel(jobID.String()),
		},
		domain.StatusRunning.String(),
		attempt,
		domain.StatusFailed.String(),
		now.UTC().Format(time.RFC3339Nano),
		jobResultPayload,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

var recoverStaleAndEnqueueScript = goredis.NewScript(`
local status = redis.call("HGET", KEYS[1], "status")

if status == ARGV[1] or status == ARGV[2] then
	return 2
end

if status == ARGV[3] then
	return 3
end

if status ~= ARGV[4] then
	return 0
end

local started_at_unix_milli = redis.call("HGET", KEYS[1], "started_at_unix_milli")
if not started_at_unix_milli then
	return 0
end

if tonumber(started_at_unix_milli) > tonumber(ARGV[5]) then
	return 0
end

redis.call("HSET", KEYS[1], "status", ARGV[3])
redis.call("HDEL", KEYS[1], "started_at", "started_at_unix_milli")
redis.call("XADD", KEYS[2], "*", "job_id", ARGV[6])

return 1
`)

func (r *RedisJobRepository) RecoverStaleAndEnqueue(ctx context.Context, jobID domain.JobID, cutoff time.Time) (domain.StaleJobRecoveryResult, error) {
	if err := jobID.Validate(); err != nil {
		return domain.StaleJobRecoveryResult{}, err
	}

	result, err := recoverStaleAndEnqueueScript.Run(
		ctx,
		r.rdb,
		[]string{
			redis.JobKey(jobID.String()),
			redis.JobStreamKey(),
		},
		domain.StatusCompleted.String(),
		domain.StatusFailed.String(),
		domain.StatusQueued.String(),
		domain.StatusRunning.String(),
		cutoff.UTC().UnixMilli(),
		jobID.String(),
	).Int()
	if err != nil {
		return domain.StaleJobRecoveryResult{}, err
	}

	return domain.StaleJobRecoveryResult{
		Recovered:     result == 1,
		Terminal:      result == 2,
		AlreadyQueued: result == 3,
	}, nil
}

func (r *RedisJobRepository) Get(ctx context.Context, jobID domain.JobID) (*domain.Job, error) {
	values, err := r.rdb.HGetAll(ctx, redis.JobKey(jobID.String())).Result()
	if err != nil {
		return nil, err
	}

	return FromRedisHash(jobID, values)
}
