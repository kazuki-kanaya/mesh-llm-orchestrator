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

func NewJob(
	jobID uuid.UUID,
	model string,
	messages json.RawMessage,
	generationParams json.RawMessage,
	now time.Time,
) *Job {
	return &Job{
		JobID:            jobID,
		Model:            model,
		Messages:         messages,
		GenerationParams: generationParams,
		Status:           StatusQueued,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}
