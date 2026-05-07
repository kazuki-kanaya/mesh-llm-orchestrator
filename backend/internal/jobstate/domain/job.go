package domain

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID             uuid.UUID
	Status         Status
	Request        HTTPRequest
	Response       *HTTPResponse
	CreatedAt      time.Time
	StartedAt      *time.Time
	TerminatedAt   *time.Time
	CurrentAttempt int64
}

func NewJob(id uuid.UUID, request HTTPRequest, now time.Time) *Job {
	return &Job{
		ID:             id,
		Status:         StatusQueued,
		Request:        request,
		CreatedAt:      now.UTC(),
		CurrentAttempt: 0,
	}
}
