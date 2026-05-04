package infrastructure

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	jobinfra "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/infrastructure"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobRepository struct {
	rdb         *goredis.Client
	terminalTTL time.Duration
}

func NewRedisJobRepository(rdb *goredis.Client, terminalTTL time.Duration) ports.JobRepository {
	if terminalTTL <= 0 {
		terminalTTL = 24 * time.Hour
	}

	return &RedisJobRepository{
		rdb:         rdb,
		terminalTTL: terminalTTL,
	}
}

var claimJobScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") == ARGV[1] then
	redis.call("HSET", KEYS[1], "status", ARGV[2], "started_at", ARGV[3])
	redis.call("ZADD", KEYS[2], ARGV[4], ARGV[5])
	return 1
end
return 0
`)

func (repo *RedisJobRepository) Claim(ctx context.Context, jobID uuid.UUID) (bool, error) {
	now := time.Now().UTC()

	result, err := claimJobScript.Run(
		ctx,
		repo.rdb,
		[]string{
			redis.JobKey(jobID),
			redis.RunningJobsKey(),
		},
		string(jobdomain.StatusQueued),
		string(jobdomain.StatusRunning),
		now.Format(time.RFC3339Nano),
		now.Unix(),
		jobID.String(),
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

	return jobinfra.FromRedisHash(jobID, values)
}

var completeJobScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") == ARGV[1] then
	redis.call("HSET", KEYS[1],
		"response", ARGV[2],
		"status", ARGV[3]
	)
	redis.call("ZREM", KEYS[2], ARGV[4])
	redis.call("EXPIRE", KEYS[1], ARGV[5])
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
		[]string{
			redis.JobKey(jobID),
			redis.RunningJobsKey(),
		},
		string(jobdomain.StatusRunning),
		responseBytes,
		string(jobdomain.StatusCompleted),
		jobID.String(),
		int(repo.terminalTTL.Seconds()),
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

var failJobScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") == ARGV[1] then
	redis.call("HSET", KEYS[1], "status", ARGV[2])
	redis.call("ZREM", KEYS[2], ARGV[3])
	redis.call("EXPIRE", KEYS[1], ARGV[4])
	return 1
end
return 0
`)

func (repo *RedisJobRepository) Fail(ctx context.Context, jobID uuid.UUID) (bool, error) {
	result, err := failJobScript.Run(
		ctx,
		repo.rdb,
		[]string{
			redis.JobKey(jobID),
			redis.RunningJobsKey(),
		},
		string(jobdomain.StatusRunning),
		string(jobdomain.StatusFailed),
		jobID.String(),
		int(repo.terminalTTL.Seconds()),
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

var _ ports.JobRepository = (*RedisJobRepository)(nil)
