package ports

import (
	"context"

	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type HTTPClient interface {
	// Do returns nil error for any HTTP response received from the upstream,
	// including 4xx and 5xx. It returns an error only when no upstream response
	// is available, such as connection failures or timeouts.
	Do(ctx context.Context, req jobdomain.HTTPRequest) (*jobdomain.HTTPResponse, error)
}
