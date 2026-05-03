package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobRepository struct {
	rdb *goredis.Client
}

func NewRedisJobRepository(rdb *goredis.Client) ports.JobRepository {
	return &RedisJobRepository{
		rdb: rdb,
	}
}

var claimJobScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") == ARGV[1] then
	redis.call("HSET", KEYS[1], "status", ARGV[2])
	return 1
end
return 0
`)

func (repo *RedisJobRepository) Claim(ctx context.Context, jobID uuid.UUID) (bool, error) {
	result, err := claimJobScript.Run(
		ctx,
		repo.rdb,
		[]string{redis.JobKey(jobID)},
		jobdomain.StatusQueued,
		jobdomain.StatusRunning,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (repo *RedisJobRepository) Get(ctx context.Context, jobID uuid.UUID) (*jobdomain.Job, error) {
	values, err := repo.rdb.HGetAll(ctx, redis.JobKey(jobID)).Result()
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	retryCount, err := strconv.Atoi(values["retry_count"])
	if err != nil {
		return nil, err
	}

	var req jobdomain.HTTPRequest
	if err := json.Unmarshal([]byte(values["request"]), &req); err != nil {
		return nil, err
	}

	var resp *jobdomain.HTTPResponse
	if rawResponse := values["response"]; rawResponse != "" {
		var decoded jobdomain.HTTPResponse
		if err := json.Unmarshal([]byte(rawResponse), &decoded); err != nil {
			return nil, err
		}
		resp = &decoded
	}

	job := &jobdomain.Job{
		ID:         jobID,
		Status:     jobdomain.Status(values["status"]),
		Request:    req,
		Response:   resp,
		RetryCount: retryCount,
	}
	return job, nil
}

var completeJobScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") == ARGV[1] then
	redis.call("HSET", KEYS[1],
		"response", ARGV[2],
		"status", ARGV[3]
	)
	return 1
end
return 0
`)

func (repo *RedisJobRepository) Complete(ctx context.Context, jobID uuid.UUID, resp jobdomain.HTTPResponse) (bool, error) {
	responseBytes, err := json.Marshal(resp)
	if err != nil {
		return false, err
	}

	result, err := completeJobScript.Run(
		ctx,
		repo.rdb,
		[]string{redis.JobKey(jobID)},
		jobdomain.StatusRunning,
		responseBytes,
		jobdomain.StatusCompleted,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

var failJobScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") == ARGV[1] then
	redis.call("HSET", KEYS[1], "status", ARGV[2])
	return 1
end
return 0
`)

func (repo *RedisJobRepository) Fail(ctx context.Context, jobID uuid.UUID) (bool, error) {
	result, err := failJobScript.Run(
		ctx,
		repo.rdb,
		[]string{redis.JobKey(jobID)},
		jobdomain.StatusRunning,
		jobdomain.StatusFailed,
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

var _ ports.JobRepository = (*RedisJobRepository)(nil)
