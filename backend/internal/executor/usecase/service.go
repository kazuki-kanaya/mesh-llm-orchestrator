package usecase

import (
	"errors"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	jobmessagingports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/ports"
)

var (
	ErrNilJobQueue           = errors.New("job queue is nil")
	ErrNilJobExecutionClient = errors.New("job execution client is nil")
	ErrNilHTTPClient         = errors.New("http client is nil")
)

type Service struct {
	queue              jobmessagingports.JobQueue
	jobExecutionClient ports.JobExecutionClient
	httpClient         ports.HTTPClient
}

func NewService(queue jobmessagingports.JobQueue, jobExecutionClient ports.JobExecutionClient, httpClient ports.HTTPClient) (*Service, error) {
	if queue == nil {
		return nil, ErrNilJobQueue
	}
	if jobExecutionClient == nil {
		return nil, ErrNilJobExecutionClient
	}
	if httpClient == nil {
		return nil, ErrNilHTTPClient
	}

	return &Service{
		queue:              queue,
		jobExecutionClient: jobExecutionClient,
		httpClient:         httpClient,
	}, nil
}
