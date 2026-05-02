package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	JobID            uuid.UUID
	Model            string
	Messages         json.RawMessage
	GenerationParams json.RawMessage
	Status           Status
	FinalResult      *string
	ErrorMessage     *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
