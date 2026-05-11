package infrastructure

import (
	"context"
	"errors"
	"time"

	jobstatev1 "github.com/kazuki-kanaya/quegress/gen/go/jobstate/v1"
	jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
	"github.com/kazuki-kanaya/quegress/internal/reconciler/ports"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var ErrNilJobStateServiceClient = errors.New("job state service client is nil")

type GRPCJobReconcileClient struct {
	jobState jobstatev1.JobStateServiceClient
}

var _ ports.JobReconcileClient = (*GRPCJobReconcileClient)(nil)

func NewGRPCJobReconcileClient(jobState jobstatev1.JobStateServiceClient) (ports.JobReconcileClient, error) {
	if jobState == nil {
		return nil, ErrNilJobStateServiceClient
	}

	return &GRPCJobReconcileClient{
		jobState: jobState,
	}, nil
}

func (c *GRPCJobReconcileClient) RecoverStaleAndEnqueue(ctx context.Context, jobID jobstatedomain.JobID, cutoff time.Time) (jobstatedomain.StaleJobRecoveryResult, error) {
	resp, err := c.jobState.RecoverStaleAndEnqueue(ctx, &jobstatev1.RecoverStaleAndEnqueueRequest{
		JobId:  jobID.String(),
		Cutoff: timestamppb.New(cutoff),
	})
	if err != nil {
		return jobstatedomain.StaleJobRecoveryResult{}, err
	}

	return jobstatedomain.StaleJobRecoveryResult{
		Recovered:     resp.GetRecovered(),
		Terminal:      resp.GetTerminal(),
		AlreadyQueued: resp.GetAlreadyQueued(),
	}, nil
}
