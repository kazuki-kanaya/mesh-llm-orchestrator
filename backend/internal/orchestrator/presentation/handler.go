package presentation

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	jobdomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/job/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/usecase"
)

type ProxyHandler struct {
	createJob     *usecase.CreateJobUseCase
	waitJob       *usecase.WaitJobUseCase
	targetBaseURL string
	waitTimeout   time.Duration
}

func NewProxyHandler(
	createJob *usecase.CreateJobUseCase,
	waitJob *usecase.WaitJobUseCase,
	targetBaseURL string,
	waitTimeout time.Duration,
) *ProxyHandler {
	return &ProxyHandler{
		createJob:     createJob,
		waitJob:       waitJob,
		targetBaseURL: targetBaseURL,
		waitTimeout:   waitTimeout,
	}
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req, err := h.buildHTTPRequest(r)
	if err != nil {
		log.Printf("failed to build job request: %v", err)
		http.Error(w, "failed to build request", http.StatusBadRequest)
		return
	}

	jobID, err := h.createJob.Execute(r.Context(), req)
	if err != nil {
		log.Printf("failed to create job: %v", err)
		http.Error(w, "failed to create job", http.StatusInternalServerError)
		return
	}

	waitCtx, cancel := context.WithTimeout(r.Context(), h.waitTimeout)
	defer cancel()

	resp, err := h.waitJob.Execute(waitCtx, jobID)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			writeAcceptedJob(w, jobID.String())
			return
		}

		log.Printf("failed to wait job: job_id=%s err=%v", jobID, err)
		http.Error(w, "failed to wait job", http.StatusBadGateway)
		return
	}

	writeUpstreamResponse(w, resp)
}

func (h *ProxyHandler) buildHTTPRequest(r *http.Request) (jobdomain.HTTPRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return jobdomain.HTTPRequest{}, err
	}
	defer r.Body.Close()

	targetURL, err := h.buildTargetURL(r)
	if err != nil {
		return jobdomain.HTTPRequest{}, err
	}

	return jobdomain.HTTPRequest{
		Method:  r.Method,
		URL:     targetURL,
		Headers: r.Header.Clone(),
		Body:    body,
	}, nil
}

func (h *ProxyHandler) buildTargetURL(r *http.Request) (string, error) {
	base, err := url.Parse(h.targetBaseURL)
	if err != nil {
		return "", err
	}

	target := *base
	target.Path = joinPath(base.Path, r.URL.Path)
	target.RawQuery = r.URL.RawQuery

	return target.String(), nil
}

func writeAcceptedJob(w http.ResponseWriter, jobID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_, _ = w.Write([]byte(`{"job_id":"` + jobID + `","status":"queued"}`))
}

func writeUpstreamResponse(w http.ResponseWriter, resp *jobdomain.HTTPResponse) {
	if resp == nil {
		http.Error(w, "upstream response is empty", http.StatusBadGateway)
		return
	}

	for key, values := range resp.Headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(resp.Body)
}

func joinPath(basePath, requestPath string) string {
	if basePath == "" {
		return requestPath
	}
	if requestPath == "" || requestPath == "/" {
		return basePath
	}
	if basePath[len(basePath)-1] == '/' && requestPath[0] == '/' {
		return basePath + requestPath[1:]
	}
	if basePath[len(basePath)-1] != '/' && requestPath[0] != '/' {
		return basePath + "/" + requestPath
	}
	return basePath + requestPath
}
