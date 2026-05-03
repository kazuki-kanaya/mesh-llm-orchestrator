package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
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
	requestBytes, err := json.Marshal(job.Request)
	if err != nil {
		return err
	}

	return repo.rdb.HSet(ctx, redis.JobKey(job.ID), map[string]any{
		"status":      job.Status,
		"request":     requestBytes,
		"retry_count": job.RetryCount,
	}).Err()
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

var _ ports.JobRepository = (*RedisJobRepository)(nil)
