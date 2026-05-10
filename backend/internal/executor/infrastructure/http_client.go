package infrastructure

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
)

var ErrNilHTTPClient = errors.New("http client is nil")

type HTTPClient struct {
	client *http.Client
}

var _ ports.HTTPClient = (*HTTPClient)(nil)

func NewHTTPClient(client *http.Client) (ports.HTTPClient, error) {
	if client == nil {
		return nil, ErrNilHTTPClient
	}

	return &HTTPClient{
		client: client,
	}, nil
}

func (c *HTTPClient) Do(ctx context.Context, req jobstatedomain.HTTPRequest) (*jobstatedomain.HTTPResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return nil, err
	}

	httpReq.Header = req.Headers.Clone()
	if req.Host != "" {
		httpReq.Host = req.Host
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &jobstatedomain.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
	}, nil
}
