package infrastructure

import (
	"encoding/json"
	"errors"
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
