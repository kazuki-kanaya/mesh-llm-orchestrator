package usecase

import (
	"context"
	"errors"

	jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
	"github.com/kazuki-kanaya/quegress/internal/requestapi/ports"
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
	jobID, err := uc.client.CreateAndEnqueue(ctx, input.Request)
	if err != nil {
		return nil, err
	}

	terminalJob, err := uc.client.Wait(ctx, jobID)
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
