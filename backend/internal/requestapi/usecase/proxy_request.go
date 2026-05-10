package usecase

import (
	"context"
	"errors"
	"time"

	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/requestapi/ports"
)

type ProxyRequestUseCase struct {
	client ports.JobRequestClient
}

var ErrNilJobRequestClient = errors.New("job request client is nil")

func NewProxyRequestUseCase(client ports.JobRequestClient) (*ProxyRequestUseCase, error) {
	if client == nil {
		return nil, ErrNilJobRequestClient
	}

	return &ProxyRequestUseCase{
		client: client,
	}, nil
}

type ProxyRequestInput struct {
	Request jobstatedomain.HTTPRequest
}

type ProxyRequestOutput struct {
	Response *jobstatedomain.HTTPResponse
}

func (uc *ProxyRequestUseCase) Execute(ctx context.Context, input ProxyRequestInput) (*ProxyRequestOutput, error) {
	id := jobstatedomain.NewJobID()
	job, err := jobstatedomain.NewJob(id, input.Request, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if err := uc.client.CreateAndEnqueue(ctx, job); err != nil {
		return nil, err
	}

	terminalJob, err := uc.client.Wait(ctx, job.ID)
	if err != nil {
		return nil, err
	}

	if terminalJob == nil {
		return nil, jobstatedomain.ErrJobNotFound
	}
	if terminalJob.Response == nil {
		return nil, jobstatedomain.ErrNilHTTPResponse
	}

	return &ProxyRequestOutput{Response: terminalJob.Response}, nil
}
