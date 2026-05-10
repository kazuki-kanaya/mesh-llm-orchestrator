package infrastructure

import (
	"context"
	"errors"

	jobmessagingports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/ports"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	jobstateports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/requestapi/ports"
)

var (
	ErrNilJobStateStore       = errors.New("job state store is nil")
	ErrNilJobResultSubscriber = errors.New("job result subscriber is nil")
)

type JobRequestClient struct {
	jobState   jobstateports.JobStateStore
	subscriber jobmessagingports.JobResultSubscriber
}

var _ ports.JobRequestClient = (*JobRequestClient)(nil)

func NewJobRequestClient(
	jobState jobstateports.JobStateStore,
	subscriber jobmessagingports.JobResultSubscriber,
) (ports.JobRequestClient, error) {
	if jobState == nil {
		return nil, ErrNilJobStateStore
	}
	if subscriber == nil {
		return nil, ErrNilJobResultSubscriber
	}

	return &JobRequestClient{
		jobState:   jobState,
		subscriber: subscriber,
	}, nil
}

func (c *JobRequestClient) CreateAndEnqueue(ctx context.Context, job *jobstatedomain.Job) error {
	return c.jobState.CreateAndEnqueue(ctx, job)
}

func (c *JobRequestClient) Wait(ctx context.Context, jobID jobstatedomain.JobID) (*jobstatedomain.Job, error) {
	if err := jobID.Validate(); err != nil {
		return nil, err
	}

	subscription, err := c.subscriber.Subscribe(ctx, jobID)
	if err != nil {
		return nil, err
	}
	defer subscription.Close()

	job, err := c.jobState.Get(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if job.Status.IsTerminal() {
		return job, nil
	}

	if err := subscription.Wait(ctx); err != nil {
		return nil, err
	}

	return c.jobState.Get(ctx, jobID)
}
