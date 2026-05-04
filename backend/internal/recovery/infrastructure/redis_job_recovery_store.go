package infrastructure

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/recovery/ports"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobRecoveryStore struct {
	rdb *goredis.Client
}

func NewRedisJobRecoveryStore(rdb *goredis.Client) ports.JobRecoveryStore {
	return &RedisJobRecoveryStore{
		rdb: rdb,
	}
}

func (s *RedisJobRecoveryStore) FindStaleRunning(ctx context.Context, cutoff time.Time, limit int64) ([]uuid.UUID, error) {
	results, err := s.rdb.ZRangeArgs(ctx, goredis.ZRangeArgs{
		Key:     redis.RunningJobsKey(),
		ByScore: true,
		Start:   "-inf",
		Stop:    strconv.FormatInt(cutoff.Unix(), 10),
		Offset:  0,
		Count:   limit,
	}).Result()
	if err != nil {
		return nil, err
	}

	jobIDs := make([]uuid.UUID, 0, len(results))
	for _, result := range results {
		jobID, err := uuid.Parse(result)
		if err != nil {
			return nil, fmt.Errorf("invalid running job id %q: %w", result, err)
		}
		jobIDs = append(jobIDs, jobID)
	}

	return jobIDs, nil
}

var recoverStaleJobScript = goredis.NewScript(`
if redis.call("HGET", KEYS[1], "status") ~= ARGV[1] then
	return 0
end

local started_at_score = redis.call("ZSCORE", KEYS[2], ARGV[3])
if not started_at_score then
	return 0
end

if tonumber(started_at_score) > tonumber(ARGV[2]) then
	return 0
end

redis.call("HSET", KEYS[1], "status", ARGV[4])
redis.call("HDEL", KEYS[1], "started_at")
redis.call("ZREM", KEYS[2], ARGV[3])
redis.call("RPUSH", KEYS[3], ARGV[3])

return 1
`)

func (s *RedisJobRecoveryStore) RecoverStale(ctx context.Context, jobID uuid.UUID, cutoff time.Time) (bool, error) {
	result, err := recoverStaleJobScript.Run(
		ctx,
		s.rdb,
		[]string{
			redis.JobKey(jobID),
			redis.RunningJobsKey(),
			redis.JobQueueKey(),
		},
		string(jobdomain.StatusRunning),
		strconv.FormatInt(cutoff.Unix(), 10),
		jobID.String(),
		string(jobdomain.StatusQueued),
	).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

var _ ports.JobRecoveryStore = (*RedisJobRecoveryStore)(nil)
