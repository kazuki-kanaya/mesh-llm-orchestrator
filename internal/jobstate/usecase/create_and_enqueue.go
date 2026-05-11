package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
	"github.com/kazuki-kanaya/quegress/internal/jobstate/ports"
)

type CreateAndEnqueueUseCase struct {
	store ports.JobStateStore
}

func NewCreateAndEnqueueUseCase(store ports.JobStateStore) (*CreateAndEnqueueUseCase, error) {
	if store == nil {
		return nil, ErrNilJobStateStore
	}

	return &CreateAndEnqueueUseCase{
		store: store,
	}, nil
}

type CreateAndEnqueueInput struct {
	Request domain.HTTPRequest
}

type CreateAndEnqueueOutput struct {
	JobID domain.JobID
}

func (uc *CreateAndEnqueueUseCase) Execute(ctx context.Context, input CreateAndEnqueueInput) (*CreateAndEnqueueOutput, error) {
	jobID := domain.NewJobID()

	job, err := domain.NewJob(jobID, input.Request, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	err = uc.store.CreateAndEnqueue(ctx, job)
	if err != nil {
		return nil, err
	}

	return &CreateAndEnqueueOutput{
		JobID: jobID,
	}, nil
}
