package ports

import (
	"context"

	jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
)

type HTTPClient interface {
	Do(ctx context.Context, req jobstatedomain.HTTPRequest) (*jobstatedomain.HTTPResponse, error)
}
