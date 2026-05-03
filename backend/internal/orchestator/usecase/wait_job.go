package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestator/ports"
)

var ErrJobNotTerminal = errors.New("job is not terminal")

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
	sub, err := uc.sub.Subscribe(ctx, jobID)
	if err != nil {
		return nil, err
	}
	defer sub.Close()

	job, err := uc.repo.Get(ctx, jobID)
	if err != nil {
		return nil, err
	}

	if job.Status.IsTerminal() {
		return job.Response, nil
	}

	select {
	case <-sub.Channel():
		job, err := uc.repo.Get(ctx, jobID)
		if err != nil {
			return nil, err
		}

		if job.Status.IsTerminal() {
			return job.Response, nil
		}

		return nil, ErrJobNotTerminal

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
