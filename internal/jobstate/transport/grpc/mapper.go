package jobstategrpc

import (
	"errors"
	"net/http"
	"time"

	jobstatev1 "github.com/kazuki-kanaya/quegress/gen/go/jobstate/v1"
	"github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrNilProtoJob           = errors.New("proto job is nil")
	ErrInvalidProtoJobStatus = errors.New("invalid proto job status")
	ErrNilProtoJobCreatedAt  = errors.New("proto job created_at is nil")
)

func HTTPRequestToProto(req domain.HTTPRequest) *jobstatev1.HTTPRequest {
	return httpRequestToProto(req)
}

func HTTPResponseToProto(resp *domain.HTTPResponse) *jobstatev1.HTTPResponse {
	return httpResponseToProto(resp)
}

func JobFromProto(job *jobstatev1.Job) (*domain.Job, error) {
	return jobFromProto(job)
}

func httpRequestFromProto(req *jobstatev1.HTTPRequest) domain.HTTPRequest {
	if req == nil {
		return domain.HTTPRequest{}
	}

	return domain.HTTPRequest{
		Method:  req.GetMethod(),
		URL:     req.GetUrl(),
		Host:    req.GetHost(),
		Headers: headersFromProto(req.GetHeaders()),
		Body:    cloneBytes(req.GetBody()),
	}
}

func httpResponseFromProto(resp *jobstatev1.HTTPResponse) *domain.HTTPResponse {
	if resp == nil {
		return nil
	}

	return &domain.HTTPResponse{
		StatusCode: int(resp.GetStatusCode()),
		Headers:    headersFromProto(resp.GetHeaders()),
		Body:       cloneBytes(resp.GetBody()),
	}
}

func httpRequestToProto(req domain.HTTPRequest) *jobstatev1.HTTPRequest {
	return &jobstatev1.HTTPRequest{
		Method:  req.Method,
		Url:     req.URL,
		Host:    req.Host,
		Headers: headersToProto(req.Headers),
		Body:    cloneBytes(req.Body),
	}
}

func httpResponseToProto(resp *domain.HTTPResponse) *jobstatev1.HTTPResponse {
	if resp == nil {
		return nil
	}

	return &jobstatev1.HTTPResponse{
		StatusCode: int32(resp.StatusCode),
		Headers:    headersToProto(resp.Headers),
		Body:       cloneBytes(resp.Body),
	}
}

func jobToProto(job *domain.Job) *jobstatev1.Job {
	if job == nil {
		return nil
	}

	return &jobstatev1.Job{
		Id:             job.ID.String(),
		Status:         job.Status.String(),
		Request:        httpRequestToProto(job.Request),
		Response:       httpResponseToProto(job.Response),
		CreatedAt:      timestamppb.New(job.CreatedAt),
		StartedAt:      timestampFromTimePtr(job.StartedAt),
		TerminatedAt:   timestampFromTimePtr(job.TerminatedAt),
		CurrentAttempt: job.CurrentAttempt,
	}
}

func jobFromProto(job *jobstatev1.Job) (*domain.Job, error) {
	if job == nil {
		return nil, ErrNilProtoJob
	}

	jobID, err := domain.ParseJobID(job.GetId())
	if err != nil {
		return nil, err
	}

	jobStatus := domain.Status(job.GetStatus())
	if !jobStatus.IsValid() {
		return nil, ErrInvalidProtoJobStatus
	}
	if job.GetCreatedAt() == nil {
		return nil, ErrNilProtoJobCreatedAt
	}

	return &domain.Job{
		ID:             jobID,
		Status:         jobStatus,
		Request:        httpRequestFromProto(job.GetRequest()),
		Response:       httpResponseFromProto(job.GetResponse()),
		CreatedAt:      job.GetCreatedAt().AsTime(),
		StartedAt:      timePtrFromTimestamp(job.GetStartedAt()),
		TerminatedAt:   timePtrFromTimestamp(job.GetTerminatedAt()),
		CurrentAttempt: job.GetCurrentAttempt(),
	}, nil
}

func headersFromProto(headers map[string]*jobstatev1.HeaderValues) http.Header {
	if len(headers) == 0 {
		return nil
	}

	result := make(http.Header, len(headers))
	for key, headerValues := range headers {
		if headerValues == nil {
			continue
		}
		result[key] = cloneStrings(headerValues.GetValues())
	}

	return result
}

func headersToProto(headers http.Header) map[string]*jobstatev1.HeaderValues {
	if len(headers) == 0 {
		return nil
	}

	result := make(map[string]*jobstatev1.HeaderValues, len(headers))
	for key, values := range headers {
		result[key] = &jobstatev1.HeaderValues{
			Values: cloneStrings(values),
		}
	}

	return result
}

func timestampFromTimePtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func timePtrFromTimestamp(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}

	value := t.AsTime()
	return &value
}

func cloneStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneBytes(values []byte) []byte {
	if len(values) == 0 {
		return nil
	}

	cloned := make([]byte, len(values))
	copy(cloned, values)
	return cloned
}
