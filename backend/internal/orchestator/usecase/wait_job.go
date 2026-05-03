package usecase

import (
	"context"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestator/ports"
)

type WaitJobUseCase struct {
	repo ports.JobRepository
	sub  ports.JobSubscriber
}

func NewWaitJobUseCase(repo ports.JobRepository, sub ports.JobSubscriber) *WaitJobUseCase {
	return &WaitJobUseCase{
		repo: repo,
		sub:  sub,
	}
}

func (uc *WaitJobUseCase) Execute(ctx context.Context, jobID uuid.UUID) (*jobdomain.HTTPResponse, error) {
	sub := uc.sub.Subscribe(ctx, jobID)
	defer sub.Close()

	job, err := uc.repo.Get(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if job.Status.IsTerminal() {
		return job.Response, nil
	}

	for {
		select {
		case <-sub.Channel():
			job, err := uc.repo.Get(ctx, jobID)
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
