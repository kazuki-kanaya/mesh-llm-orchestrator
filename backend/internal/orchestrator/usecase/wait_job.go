package usecase

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/ports"
)

type WaitJobUseCase struct {
	reader ports.JobReader
	sub    ports.JobSubscriber
}

func NewWaitJobUseCase(reader ports.JobReader, sub ports.JobSubscriber) *WaitJobUseCase {
	return &WaitJobUseCase{
		reader: reader,
		sub:    sub,
	}
}

func (uc *WaitJobUseCase) Execute(ctx context.Context, jobID uuid.UUID) (*jobdomain.HTTPResponse, error) {
	sub, err := uc.sub.Subscribe(ctx, jobID)
	if err != nil {
		return nil, err
	}
	defer sub.Close()

	job, err := uc.reader.Get(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if job.Status.IsTerminal() {
		return job.Response, nil
	}

	for {
		select {
		case <-sub.Channel():
			job, err := uc.reader.Get(ctx, jobID)
			if err != nil {
				return nil, err
			}

			if job.Status.IsTerminal() {
				return job.Response, nil
			}

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
