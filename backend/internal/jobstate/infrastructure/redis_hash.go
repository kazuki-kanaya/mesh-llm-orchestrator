package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

var ErrNilJob = errors.New("job is nil")

func ToRedisHash(job *domain.Job) (map[string]any, error) {
	if job == nil {
		return nil, ErrNilJob
	}

	requestBytes, err := json.Marshal(job.Request)
	if err != nil {
		return nil, err
	}

	values := map[string]any{
		"status":          job.Status.String(),
		"request":         requestBytes,
		"created_at":      job.CreatedAt.Format(time.RFC3339Nano),
		"current_attempt": job.CurrentAttempt,
	}

	if job.StartedAt != nil {
		values["started_at"] = job.StartedAt.Format(time.RFC3339Nano)
	}

	if job.TerminatedAt != nil {
		values["terminated_at"] = job.TerminatedAt.Format(time.RFC3339Nano)
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

func FromRedisHash(jobID domain.JobID, values map[string]string) (*domain.Job, error) {
	if len(values) == 0 {
		return nil, fmt.Errorf("%w: %s", domain.ErrJobNotFound, jobID)
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

	var req domain.HTTPRequest
	if err := json.Unmarshal([]byte(rawRequest), &req); err != nil {
		return nil, fmt.Errorf("invalid job request: %w", err)
	}

	rawCreatedAt := values["created_at"]
	if rawCreatedAt == "" {
		return nil, fmt.Errorf("missing job created_at: %s", jobID)
	}
	createdAt, err := time.Parse(time.RFC3339Nano, rawCreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid job created_at %q: %w", rawCreatedAt, err)
	}

	rawCurrentAttempt := values["current_attempt"]
	if rawCurrentAttempt == "" {
		return nil, fmt.Errorf("missing job current_attempt: %s", jobID)
	}
	currentAttempt, err := strconv.ParseInt(rawCurrentAttempt, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid job current_attempt %q: %w", rawCurrentAttempt, err)
	}

	var startedAt *time.Time
	if rawStartedAt := values["started_at"]; rawStartedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, rawStartedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid job started_at %q: %w", rawStartedAt, err)
		}
		startedAt = &parsed
	}

	var terminatedAt *time.Time
	if rawTerminatedAt := values["terminated_at"]; rawTerminatedAt != "" {
		parsed, err := time.Parse(time.RFC3339Nano, rawTerminatedAt)
		if err != nil {
			return nil, fmt.Errorf("invalid job terminated_at %q: %w", rawTerminatedAt, err)
		}
		terminatedAt = &parsed
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
		ID:             jobID,
		Status:         status,
		Request:        req,
		Response:       resp,
		CreatedAt:      createdAt,
		StartedAt:      startedAt,
		TerminatedAt:   terminatedAt,
		CurrentAttempt: currentAttempt,
	}, nil
}
