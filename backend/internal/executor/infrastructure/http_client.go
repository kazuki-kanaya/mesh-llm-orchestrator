package infrastructure

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/ports"
	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
)

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient(client *http.Client) ports.HTTPClient {
	if client == nil {
		client = http.DefaultClient
	}

	return &HTTPClient{
		client: client,
	}
}

func (c *HTTPClient) Do(ctx context.Context, req jobdomain.HTTPRequest) (*jobdomain.HTTPResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(req.Body))
	if err != nil {
		return nil, err
	}

	httpReq.Header = req.Headers.Clone()
	httpReq.Host = req.Host

	httpResp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	return &jobdomain.HTTPResponse{
		StatusCode: httpResp.StatusCode,
		Headers:    httpResp.Header.Clone(),
		Body:       body,
	}, nil
}

var _ ports.HTTPClient = (*HTTPClient)(nil)
