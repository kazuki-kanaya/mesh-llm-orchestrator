package domain

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID         uuid.UUID
	Status     Status
	Request    HTTPRequest
	Response   *HTTPResponse
	StartedAt  *time.Time
	RetryCount int
}

func NewJob(id uuid.UUID, request HTTPRequest) *Job {
	return &Job{
		ID:         id,
		Status:     StatusQueued,
		Request:    request,
		RetryCount: 0,
	}
}
