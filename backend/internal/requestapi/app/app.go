package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	jobstatev1 "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/gen/go/jobstate/v1"
	jobmessaginginfra "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/infrastructure"
	platformredis "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	requestapiinfra "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/requestapi/infrastructure"
	requestapihttp "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/requestapi/transport/http"
	requestapiusecase "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/requestapi/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

var (
	ErrEmptyRedisAddr       = errors.New("redis addr is empty")
	ErrEmptyJobStateAddress = errors.New("jobstate grpc addr is empty")
	ErrEmptyHTTPAddr        = errors.New("http addr is empty")
	ErrEmptyTargetBaseURL   = errors.New("target base url is empty")
	ErrInvalidWaitTimeout   = errors.New("wait timeout must be positive")
)

type Config struct {
	RedisAddr        string
	JobStateGRPCAddr string
	HTTPAddr         string
	TargetBaseURL    string
	WaitTimeout      time.Duration
}

func Run(ctx context.Context, cfg Config) error {
	if cfg.RedisAddr == "" {
		return ErrEmptyRedisAddr
	}
	if cfg.JobStateGRPCAddr == "" {
		return ErrEmptyJobStateAddress
	}
	if cfg.HTTPAddr == "" {
		return ErrEmptyHTTPAddr
	}
	if cfg.TargetBaseURL == "" {
		return ErrEmptyTargetBaseURL
	}
	if cfg.WaitTimeout <= 0 {
		return ErrInvalidWaitTimeout
	}

	rdb, err := platformredis.NewClient(ctx, platformredis.Config{
		Addr: cfg.RedisAddr,
	})
	if err != nil {
		return err
	}
	defer rdb.Close()

	conn, err := grpc.NewClient(cfg.JobStateGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	subscriber, err := jobmessaginginfra.NewRedisResultSubscriber(rdb)
	if err != nil {
		return err
	}

	client, err := requestapiinfra.NewGRPCJobRequestClient(jobstatev1.NewJobStateServiceClient(conn), subscriber)
	if err != nil {
		return err
	}

	proxyRequest, err := requestapiusecase.NewProxyRequestUseCase(client)
	if err != nil {
		return err
	}

	handler, err := requestapihttp.NewProxyHandler(proxyRequest, cfg.TargetBaseURL, cfg.WaitTimeout)
	if err != nil {
		return err
	}

	writeTimeout := cfg.WaitTimeout + 5*time.Second
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), writeTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("failed to shutdown requestapi: %v", err)
		}
	}()

	log.Printf("requestapi listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
