package usecase

import (
	"errors"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	jobqueueports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/ports"
)

var (
	ErrNilJobQueue   = errors.New("job queue is nil")
	ErrNilJobState   = errors.New("job state is nil")
	ErrNilHTTPClient = errors.New("http client is nil")
)

type Service struct {
	queue    jobqueueports.JobQueue
	jobState ports.JobState
	client   ports.HTTPClient
}

func NewService(queue jobqueueports.JobQueue, jobState ports.JobState, client ports.HTTPClient) (*Service, error) {
	if queue == nil {
		return nil, ErrNilJobQueue
	}
	if jobState == nil {
		return nil, ErrNilJobState
	}
	if client == nil {
		return nil, ErrNilHTTPClient
	}

	return &Service{
		queue:    queue,
		jobState: jobState,
		client:   client,
	}, nil
}
