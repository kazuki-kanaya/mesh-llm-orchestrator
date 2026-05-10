package presentation

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/requestapi/usecase"
)

var (
	ErrNilProxyRequestUseCase = errors.New("proxy request usecase is nil")
	ErrInvalidWaitTimeout     = errors.New("wait timeout must be positive")
)

type ProxyHandler struct {
	proxyRequest *usecase.ProxyRequestUseCase
	targetBase   *url.URL
	waitTimeout  time.Duration
}

func NewProxyHandler(proxyRequest *usecase.ProxyRequestUseCase, targetBaseURL string, waitTimeout time.Duration) (*ProxyHandler, error) {
	if proxyRequest == nil {
		return nil, ErrNilProxyRequestUseCase
	}
	if waitTimeout <= 0 {
		return nil, ErrInvalidWaitTimeout
	}

	targetBase, err := url.Parse(targetBaseURL)
	if err != nil {
		return nil, err
	}
	if targetBase.Scheme == "" || targetBase.Host == "" {
		return nil, errors.New("target base url must include scheme and host")
	}

	return &ProxyHandler{
		proxyRequest: proxyRequest,
		targetBase:   targetBase,
		waitTimeout:  waitTimeout,
	}, nil
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := h.buildHTTPRequest(r)
	if err != nil {
		http.Error(w, "failed to build request", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.waitTimeout)
	defer cancel()

	output, err := h.proxyRequest.Execute(ctx, usecase.ProxyRequestInput{
		Request: req,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	if output == nil || output.Response == nil {
		http.Error(w, "upstream response is unavailable", http.StatusBadGateway)
		return
	}

	writeUpstreamResponse(w, output.Response)
}

func (h *ProxyHandler) buildHTTPRequest(r *http.Request) (jobstatedomain.HTTPRequest, error) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return jobstatedomain.HTTPRequest{}, err
	}

	headers := r.Header.Clone()
	removeHopByHopHeaders(headers)

	return jobstatedomain.HTTPRequest{
		Method:  r.Method,
		URL:     h.buildTargetURL(r),
		Headers: headers,
		Body:    body,
	}, nil
}

func (h *ProxyHandler) buildTargetURL(r *http.Request) string {
	target := *h.targetBase
	target.Path = joinPath(h.targetBase.Path, r.URL.Path)
	target.RawQuery = joinQuery(h.targetBase.RawQuery, r.URL.RawQuery)

	return target.String()
}

func writeUpstreamResponse(w http.ResponseWriter, resp *jobstatedomain.HTTPResponse) {
	if resp == nil {
		http.Error(w, "upstream response is unavailable", http.StatusBadGateway)
		return
	}

	headers := resp.Headers.Clone()
	removeHopByHopHeaders(headers)

	for key, values := range headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(resp.Body)
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		http.Error(w, "upstream request timed out", http.StatusGatewayTimeout)
		return
	}
	if errors.Is(err, jobstatedomain.ErrJobNotFound) || errors.Is(err, jobstatedomain.ErrNilHTTPResponse) {
		http.Error(w, "upstream response is unavailable", http.StatusBadGateway)
		return
	}

	http.Error(w, "failed to proxy request", http.StatusInternalServerError)
}

func joinPath(basePath, requestPath string) string {
	if basePath == "" {
		return requestPath
	}
	if requestPath == "" || requestPath == "/" {
		return basePath
	}
	if strings.HasSuffix(basePath, "/") && strings.HasPrefix(requestPath, "/") {
		return basePath + strings.TrimPrefix(requestPath, "/")
	}
	if !strings.HasSuffix(basePath, "/") && !strings.HasPrefix(requestPath, "/") {
		return basePath + "/" + requestPath
	}
	return basePath + requestPath
}

func joinQuery(baseQuery, requestQuery string) string {
	switch {
	case baseQuery == "":
		return requestQuery
	case requestQuery == "":
		return baseQuery
	default:
		return baseQuery + "&" + requestQuery
	}
}

func removeHopByHopHeaders(headers http.Header) {
	for _, header := range headers.Values("Connection") {
		for _, field := range strings.Split(header, ",") {
			if field = strings.TrimSpace(field); field != "" {
				headers.Del(field)
			}
		}
	}

	for _, header := range hopByHopHeaders {
		headers.Del(header)
	}
}

var hopByHopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}
