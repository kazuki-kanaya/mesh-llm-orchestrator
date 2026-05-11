package domain

import (
	"time"
)

type Job struct {
	ID             JobID
	Status         Status
	Request        HTTPRequest
	Response       *HTTPResponse
	CreatedAt      time.Time
	StartedAt      *time.Time
	TerminatedAt   *time.Time
	CurrentAttempt int64
}

func NewJob(id JobID, request HTTPRequest, now time.Time) (*Job, error) {
	if err := id.Validate(); err != nil {
		return nil, err
	}
	return &Job{
		ID:             id,
		Status:         StatusQueued,
		Request:        request,
		CreatedAt:      now.UTC(),
		CurrentAttempt: 0,
	}, nil
}
