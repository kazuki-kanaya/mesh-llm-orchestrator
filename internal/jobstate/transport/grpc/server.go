package jobstategrpc

import (
	"context"
	"errors"

	jobstatev1 "github.com/kazuki-kanaya/quegress/gen/go/jobstate/v1"
	"github.com/kazuki-kanaya/quegress/internal/jobstate/domain"
	"github.com/kazuki-kanaya/quegress/internal/jobstate/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	jobstatev1.UnimplementedJobStateServiceServer

	createAndEnqueue       *usecase.CreateAndEnqueueUseCase
	startAttempt           *usecase.StartAttemptUseCase
	completeAttempt        *usecase.CompleteAttemptUseCase
	failAttempt            *usecase.FailAttemptUseCase
	recoverStaleAndEnqueue *usecase.RecoverStaleAndEnqueueUseCase
	get                    *usecase.GetUseCase
}

func NewServer(
	createAndEnqueue *usecase.CreateAndEnqueueUseCase,
	startAttempt *usecase.StartAttemptUseCase,
	completeAttempt *usecase.CompleteAttemptUseCase,
	failAttempt *usecase.FailAttemptUseCase,
	recoverStaleAndEnqueue *usecase.RecoverStaleAndEnqueueUseCase,
	get *usecase.GetUseCase,
) (*Server, error) {
	if createAndEnqueue == nil {
		return nil, errors.New("create and enqueue usecase is nil")
	}
	if startAttempt == nil {
		return nil, errors.New("start attempt usecase is nil")
	}
	if completeAttempt == nil {
		return nil, errors.New("complete attempt usecase is nil")
	}
	if failAttempt == nil {
		return nil, errors.New("fail attempt usecase is nil")
	}
	if recoverStaleAndEnqueue == nil {
		return nil, errors.New("recover stale and enqueue usecase is nil")
	}
	if get == nil {
		return nil, errors.New("get usecase is nil")
	}

	return &Server{
		createAndEnqueue:       createAndEnqueue,
		startAttempt:           startAttempt,
		completeAttempt:        completeAttempt,
		failAttempt:            failAttempt,
		recoverStaleAndEnqueue: recoverStaleAndEnqueue,
		get:                    get,
	}, nil
}

func (s *Server) CreateAndEnqueue(ctx context.Context, req *jobstatev1.CreateAndEnqueueRequest) (*jobstatev1.CreateAndEnqueueResponse, error) {
	if req == nil || req.GetRequest() == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	output, err := s.createAndEnqueue.Execute(ctx, usecase.CreateAndEnqueueInput{
		Request: httpRequestFromProto(req.GetRequest()),
	})
	if err != nil {
		return nil, toStatusError(err)
	}

	return &jobstatev1.CreateAndEnqueueResponse{
		JobId: output.JobID.String(),
	}, nil
}

func (s *Server) StartAttempt(ctx context.Context, req *jobstatev1.StartAttemptRequest) (*jobstatev1.StartAttemptResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	jobID, err := parseJobID(req.GetJobId())
	if err != nil {
		return nil, toStatusError(err)
	}

	output, err := s.startAttempt.Execute(ctx, usecase.StartAttemptInput{
		JobID: jobID,
	})
	if err != nil {
		return nil, toStatusError(err)
	}

	return &jobstatev1.StartAttemptResponse{
		Accepted: output.Accepted,
		Attempt:  output.Attempt,
	}, nil
}

func (s *Server) CompleteAttempt(ctx context.Context, req *jobstatev1.CompleteAttemptRequest) (*jobstatev1.CompleteAttemptResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.GetResponse() == nil {
		return nil, status.Error(codes.InvalidArgument, "response is nil")
	}

	jobID, err := parseJobID(req.GetJobId())
	if err != nil {
		return nil, toStatusError(err)
	}

	output, err := s.completeAttempt.Execute(ctx, usecase.CompleteAttemptInput{
		JobID:    jobID,
		Attempt:  req.GetAttempt(),
		Response: httpResponseFromProto(req.GetResponse()),
	})
	if err != nil {
		return nil, toStatusError(err)
	}

	return &jobstatev1.CompleteAttemptResponse{
		Accepted: output.Accepted,
	}, nil
}

func (s *Server) FailAttempt(ctx context.Context, req *jobstatev1.FailAttemptRequest) (*jobstatev1.FailAttemptResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	jobID, err := parseJobID(req.GetJobId())
	if err != nil {
		return nil, toStatusError(err)
	}

	output, err := s.failAttempt.Execute(ctx, usecase.FailAttemptInput{
		JobID:   jobID,
		Attempt: req.GetAttempt(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}

	return &jobstatev1.FailAttemptResponse{
		Accepted: output.Accepted,
	}, nil
}

func (s *Server) RecoverStaleAndEnqueue(ctx context.Context, req *jobstatev1.RecoverStaleAndEnqueueRequest) (*jobstatev1.RecoverStaleAndEnqueueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	if req.GetCutoff() == nil {
		return nil, status.Error(codes.InvalidArgument, "cutoff is nil")
	}

	jobID, err := parseJobID(req.GetJobId())
	if err != nil {
		return nil, toStatusError(err)
	}

	output, err := s.recoverStaleAndEnqueue.Execute(ctx, usecase.RecoverStaleAndEnqueueInput{
		JobID:  jobID,
		Cutoff: req.GetCutoff().AsTime(),
	})
	if err != nil {
		return nil, toStatusError(err)
	}

	return &jobstatev1.RecoverStaleAndEnqueueResponse{
		Recovered:     output.Recovered,
		Terminal:      output.Terminal,
		AlreadyQueued: output.AlreadyQueued,
	}, nil
}

func (s *Server) Get(ctx context.Context, req *jobstatev1.GetRequest) (*jobstatev1.GetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	jobID, err := parseJobID(req.GetJobId())
	if err != nil {
		return nil, toStatusError(err)
	}

	output, err := s.get.Execute(ctx, usecase.GetInput{
		JobID: jobID,
	})
	if err != nil {
		return nil, toStatusError(err)
	}
	if output.Job == nil {
		return nil, toStatusError(domain.ErrJobNotFound)
	}

	return &jobstatev1.GetResponse{
		Job: jobToProto(output.Job),
	}, nil
}

func parseJobID(value string) (domain.JobID, error) {
	return domain.ParseJobID(value)
}

func toStatusError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrInvalidJobID):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNilHTTPResponse):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrJobNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
