package infrastructure

import (
	"context"
	"errors"

	jobstatev1 "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/gen/go/jobstate/v1"
	jobmessagingports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/ports"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	jobstategrpc "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/transport/grpc"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/requestapi/ports"
)

var (
	ErrNilJobStateServiceClient = errors.New("job state service client is nil")
	ErrNilJobResultSubscriber   = errors.New("job result subscriber is nil")
)

type GRPCJobRequestClient struct {
	jobState   jobstatev1.JobStateServiceClient
	subscriber jobmessagingports.JobResultSubscriber
}

var _ ports.JobRequestClient = (*GRPCJobRequestClient)(nil)

func NewGRPCJobRequestClient(
	jobState jobstatev1.JobStateServiceClient,
	subscriber jobmessagingports.JobResultSubscriber,
) (ports.JobRequestClient, error) {
	if jobState == nil {
		return nil, ErrNilJobStateServiceClient
	}
	if subscriber == nil {
		return nil, ErrNilJobResultSubscriber
	}

	return &GRPCJobRequestClient{
		jobState:   jobState,
		subscriber: subscriber,
	}, nil
}

func (c *GRPCJobRequestClient) CreateAndEnqueue(ctx context.Context, request jobstatedomain.HTTPRequest) (jobstatedomain.JobID, error) {
	var zero jobstatedomain.JobID

	resp, err := c.jobState.CreateAndEnqueue(ctx, &jobstatev1.CreateAndEnqueueRequest{
		Request: jobstategrpc.HTTPRequestToProto(request),
	})
	if err != nil {
		return zero, err
	}

	jobID, err := jobstatedomain.ParseJobID(resp.GetJobId())
	if err != nil {
		return zero, err
	}

	return jobID, nil
}

func (c *GRPCJobRequestClient) Wait(ctx context.Context, jobID jobstatedomain.JobID) (*jobstatedomain.Job, error) {
	if err := jobID.Validate(); err != nil {
		return nil, err
	}

	subscription, err := c.subscriber.Subscribe(ctx, jobID)
	if err != nil {
		return nil, err
	}
	defer subscription.Close()

	job, err := c.get(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status.IsTerminal() {
		return job, nil
	}

	if err := subscription.Wait(ctx); err != nil {
		return nil, err
	}

	return c.get(ctx, jobID)
}

func (c *GRPCJobRequestClient) get(ctx context.Context, jobID jobstatedomain.JobID) (*jobstatedomain.Job, error) {
	resp, err := c.jobState.Get(ctx, &jobstatev1.GetRequest{
		JobId: jobID.String(),
	})
	if err != nil {
		return nil, err
	}

	return jobstategrpc.JobFromProto(resp.GetJob())
}
