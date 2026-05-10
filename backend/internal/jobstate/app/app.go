package app

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	jobstatev1 "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/gen/go/jobstate/v1"
	jobstateinfra "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/infrastructure"
	jobstateports "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/ports"
	jobstategrpc "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/transport/grpc"
	jobstateusecase "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/usecase"
	platformredis "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	"google.golang.org/grpc"
)

const gracefulStopTimeout = 5 * time.Second

var (
	ErrEmptyRedisAddr = errors.New("redis addr is empty")
	ErrEmptyGRPCAddr  = errors.New("grpc addr is empty")
)

type Config struct {
	RedisAddr string
	GRPCAddr  string
}

func Run(ctx context.Context, cfg Config) error {
	if cfg.RedisAddr == "" {
		return ErrEmptyRedisAddr
	}
	if cfg.GRPCAddr == "" {
		return ErrEmptyGRPCAddr
	}

	rdb, err := platformredis.NewClient(ctx, platformredis.Config{
		Addr: cfg.RedisAddr,
	})
	if err != nil {
		return err
	}
	defer rdb.Close()

	store, err := jobstateinfra.NewRedisJobStore(rdb)
	if err != nil {
		return err
	}

	jobStateServer, err := newGRPCServer(store)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return err
	}

	server := grpc.NewServer()
	jobstatev1.RegisterJobStateServiceServer(server, jobStateServer)

	go func() {
		<-ctx.Done()
		gracefulStop(server)
	}()

	log.Printf("jobstate grpc listening on %s", listener.Addr())
	return server.Serve(listener)
}

func gracefulStop(server *grpc.Server) {
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(gracefulStopTimeout):
		server.Stop()
	}
}

func newGRPCServer(store jobstateports.JobStateStore) (*jobstategrpc.Server, error) {
	createAndEnqueue, err := jobstateusecase.NewCreateAndEnqueueUseCase(store)
	if err != nil {
		return nil, err
	}
	startAttempt, err := jobstateusecase.NewStartAttemptUseCase(store)
	if err != nil {
		return nil, err
	}
	completeAttempt, err := jobstateusecase.NewCompleteAttemptUseCase(store)
	if err != nil {
		return nil, err
	}
	failAttempt, err := jobstateusecase.NewFailAttemptUseCase(store)
	if err != nil {
		return nil, err
	}
	recoverStaleAndEnqueue, err := jobstateusecase.NewRecoverStaleAndEnqueueUseCase(store)
	if err != nil {
		return nil, err
	}
	get, err := jobstateusecase.NewGetUseCase(store)
	if err != nil {
		return nil, err
	}

	return jobstategrpc.NewServer(
		createAndEnqueue,
		startAttempt,
		completeAttempt,
		failAttempt,
		recoverStaleAndEnqueue,
		get,
	)
}
