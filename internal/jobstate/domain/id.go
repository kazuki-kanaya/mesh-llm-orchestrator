package domain

import (
	"errors"

	"github.com/google/uuid"
)

type JobID uuid.UUID

func NewJobID() JobID {
	return JobID(uuid.New())
}

func ParseJobID(value string) (JobID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return JobID(uuid.Nil), ErrInvalidJobID
	}
	return JobID(id), nil
}

func (id JobID) String() string {
	return uuid.UUID(id).String()
}

var ErrInvalidJobID = errors.New("invalid job id")

func (id JobID) Validate() error {
	if uuid.UUID(id) == uuid.Nil {
		return ErrInvalidJobID
	}
	return nil
}
