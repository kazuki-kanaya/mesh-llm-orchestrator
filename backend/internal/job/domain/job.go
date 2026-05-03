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
