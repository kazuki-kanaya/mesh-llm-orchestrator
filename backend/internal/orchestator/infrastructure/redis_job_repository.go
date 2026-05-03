package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/redis/go-redis/v9"
)

type RedisJobRepository struct {
	rdb *redis.Client
}

func NewRedisJobRepository(rdb *redis.Client) *RedisJobRepository {
	return &RedisJobRepository{
		rdb: rdb,
	}
}

func (repo *RedisJobRepository) Create(ctx context.Context, job *jobdomain.Job) error {
	requestBytes, err := json.Marshal(job.Request)
	if err != nil {
		return err
	}

	return repo.rdb.HSet(ctx, jobKey(job.ID), map[string]any{
		"status":      job.Status,
		"request":     requestBytes,
		"retry_count": 0,
	}).Err()
}

func (repo *RedisJobRepository) Get(ctx context.Context, jobID uuid.UUID) (*jobdomain.Job, error) {
	values, err := repo.rdb.HGetAll(ctx, jobKey(jobID)).Result()
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

	job := &jobdomain.Job{
		ID:         jobID,
		Status:     jobdomain.Status(values["status"]),
		Request:    req,
		RetryCount: retryCount,
	}
	return job, nil
}

func jobKey(jobID uuid.UUID) string {
	return "job:" + jobID.String()
}
