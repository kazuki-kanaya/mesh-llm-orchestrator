package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

func ToRedisHash(job *domain.Job) (map[string]any, error) {
	if job == nil {
		return nil, errors.New("job is nil")
	}

	requestBytes, err := json.Marshal(job.Request)
	if err != nil {
		return nil, err
	}

	values := map[string]any{
		"status":      string(job.Status),
		"request":     requestBytes,
		"retry_count": job.RetryCount,
	}

	if job.StartedAt != nil {
		values["started_at"] = job.StartedAt.Format(time.RFC3339Nano)
	}

	if job.Response != nil {
		responseBytes, err := json.Marshal(job.Response)
		if err != nil {
			return nil, err
		}
		values["response"] = responseBytes
	}

	return values, nil
}

func FromRedisHash(jobID uuid.UUID, values map[string]string) (*domain.Job, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	rawStatus := values["status"]
	if rawStatus == "" {
		return nil, fmt.Errorf("missing job status: %s", jobID)
	}
	status := domain.Status(rawStatus)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid job status %q: %s", rawStatus, jobID)
	}

	rawRequest := values["request"]
	if rawRequest == "" {
		return nil, fmt.Errorf("missing job request: %s", jobID)
	}

	rawRetryCount := values["retry_count"]
	if rawRetryCount == "" {
		return nil, fmt.Errorf("missing job retry_count: %s", jobID)
	}

	retryCount, err := strconv.Atoi(rawRetryCount)
	if err != nil {
		return nil, fmt.Errorf("invalid job retry_count %q: %w", rawRetryCount, err)
	}

	var req domain.HTTPRequest
	if err := json.Unmarshal([]byte(rawRequest), &req); err != nil {
		return nil, fmt.Errorf("invalid job request: %w", err)
	}

	var startedAt *time.Time
	if rawStartedAt := values["started_at"]; rawStartedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, rawStartedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid job started_at %q: %w", rawStartedAt, err)
		}
		startedAt = &parsed
	}

	var resp *domain.HTTPResponse
	if rawResponse := values["response"]; rawResponse != "" {
		var decoded domain.HTTPResponse
		if err := json.Unmarshal([]byte(rawResponse), &decoded); err != nil {
			return nil, fmt.Errorf("invalid job response: %w", err)
		}
		resp = &decoded
	}

	return &domain.Job{
		ID:         jobID,
		Status:     status,
		Request:    req,
		Response:   resp,
		StartedAt:  startedAt,
		RetryCount: retryCount,
	}, nil
}
