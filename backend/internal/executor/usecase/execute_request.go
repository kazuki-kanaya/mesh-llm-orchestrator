package usecase

import (
	"context"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/domain"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

func (s *Service) executeRequest(ctx context.Context, req jobstatedomain.HTTPRequest) (*jobstatedomain.HTTPResponse, error) {
	resp, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, domain.ErrNilHTTPResponse
	}
	return resp, nil
}
