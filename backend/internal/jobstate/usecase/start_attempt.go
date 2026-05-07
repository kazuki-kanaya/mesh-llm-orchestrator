package usecase

import (
	"context"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
)

type StartAttemptUseCase struct {
	repo ports.JobRepository
}

func NewStartAttemptUseCase(repo ports.JobRepository) *StartAttemptUseCase {
	return &StartAttemptUseCase{
		repo: repo,
	}
}

type StartAttemptRequest struct {
	JobID domain.JobID
}

type StartAttemptResponse struct {
	Accepted bool
	Attempt  int64
}

func (uc *StartAttemptUseCase) Execute(ctx context.Context, request StartAttemptRequest) (*StartAttemptResponse, error) {
	if err := request.JobID.Validate(); err != nil {
		return nil, err
	}

	accepted, attempt, err := uc.repo.StartAttempt(ctx, request.JobID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &StartAttemptResponse{
		Accepted: accepted,
		Attempt:  attempt,
	}, nil
}
