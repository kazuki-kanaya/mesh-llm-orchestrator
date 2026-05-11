package infrastructure

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/kazuki-kanaya/quegress/internal/executor/ports"
	jobstatedomain "github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
)

var (
	ErrNilHTTPClient            = errors.New("http client is nil")
	ErrInvalidMaxResponseBytes  = errors.New("max response bytes must be positive")
	ErrUpstreamResponseTooLarge = errors.New("upstream response body is too large")
)

type HTTPClient struct {
	client           *http.Client
	maxResponseBytes int64
}

var _ ports.HTTPClient = (*HTTPClient)(nil)

func NewHTTPClient(client *http.Client, maxResponseBytes int64) (ports.HTTPClient, error) {
	if client == nil {
		return nil, ErrNilHTTPClient
	}
	if maxResponseBytes <= 0 {
		return nil, ErrInvalidMaxResponseBytes
	}

	return &HTTPClient{
		client:           client,
		maxResponseBytes: maxResponseBytes,
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

	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > c.maxResponseBytes {
		return nil, ErrUpstreamResponseTooLarge
	}

	return &jobstatedomain.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header.Clone(),
		Body:       body,
	}, nil
}
