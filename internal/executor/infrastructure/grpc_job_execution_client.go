package infrastructure

import (
	"context"
	"errors"

	jobstatev1 "github.com/kazuki-kanaya/quegress/gen/go/jobstate/v1"
	"github.com/kazuki-kanaya/quegress/internal/executor/ports"
	jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
	jobstategrpc "github.com/kazuki-kanaya/quegress/internal/jobstate/transport/grpc"
)

var ErrNilJobStateServiceClient = errors.New("job state service client is nil")

type GRPCJobExecutionClient struct {
	jobState jobstatev1.JobStateServiceClient
}

var _ ports.JobExecutionClient = (*GRPCJobExecutionClient)(nil)

func NewGRPCJobExecutionClient(jobState jobstatev1.JobStateServiceClient) (ports.JobExecutionClient, error) {
	if jobState == nil {
		return nil, ErrNilJobStateServiceClient
	}

	return &GRPCJobExecutionClient{
		jobState: jobState,
	}, nil
}

func (c *GRPCJobExecutionClient) ClaimAttempt(ctx context.Context, jobID jobstatedomain.JobID) (bool, int64, error) {
	resp, err := c.jobState.StartAttempt(ctx, &jobstatev1.StartAttemptRequest{
		JobId: jobID.String(),
	})
	if err != nil {
		return false, 0, err
	}

	return resp.GetAccepted(), resp.GetAttempt(), nil
}

func (c *GRPCJobExecutionClient) Get(ctx context.Context, jobID jobstatedomain.JobID) (*jobstatedomain.Job, error) {
	resp, err := c.jobState.Get(ctx, &jobstatev1.GetRequest{
		JobId: jobID.String(),
	})
	if err != nil {
		return nil, err
	}

	return jobstategrpc.JobFromProto(resp.GetJob())
}

func (c *GRPCJobExecutionClient) CompleteAttempt(ctx context.Context, jobID jobstatedomain.JobID, attempt int64, response *jobstatedomain.HTTPResponse) (bool, error) {
	if response == nil {
		return false, jobstatedomain.ErrNilHTTPResponse
	}

	resp, err := c.jobState.CompleteAttempt(ctx, &jobstatev1.CompleteAttemptRequest{
		JobId:    jobID.String(),
		Attempt:  attempt,
		Response: jobstategrpc.HTTPResponseToProto(response),
	})
	if err != nil {
		return false, err
	}

	return resp.GetAccepted(), nil
}

func (c *GRPCJobExecutionClient) FailAttempt(ctx context.Context, jobID jobstatedomain.JobID, attempt int64) (bool, error) {
	resp, err := c.jobState.FailAttempt(ctx, &jobstatev1.FailAttemptRequest{
		JobId:   jobID.String(),
		Attempt: attempt,
	})
	if err != nil {
		return false, err
	}

	return resp.GetAccepted(), nil
}
