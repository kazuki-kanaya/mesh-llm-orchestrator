package usecase

import (
	"errors"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	jobqueueports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/ports"
)

var (
	ErrNilJobQueue     = errors.New("job queue is nil")
	ErrNilJobExecution = errors.New("job execution store is nil")
	ErrNilHTTPClient   = errors.New("http client is nil")
)

type Service struct {
	queue     jobqueueports.JobQueue
	execution ports.JobExecutionStore
	client    ports.HTTPClient
}

func NewService(queue jobqueueports.JobQueue, execution ports.JobExecutionStore, client ports.HTTPClient) (*Service, error) {
	if queue == nil {
		return nil, ErrNilJobQueue
	}
	if execution == nil {
		return nil, ErrNilJobExecution
	}
	if client == nil {
		return nil, ErrNilHTTPClient
	}

	return &Service{
		queue:     queue,
		execution: execution,
		client:    client,
	}, nil
}
