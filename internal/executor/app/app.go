package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	jobstatev1 "github.com/kazuki-kanaya/quegress/gen/go/jobstate/v1"
	executorinfra "github.com/kazuki-kanaya/quegress/internal/executor/infrastructure"
	executorusecase "github.com/kazuki-kanaya/quegress/internal/executor/usecase"
	jobmessagingdomain "github.com/kazuki-kanaya/quegress/internal/jobmessaging/domain"
	jobmessaginginfra "github.com/kazuki-kanaya/quegress/internal/jobmessaging/infrastructure"
	platformredis "github.com/kazuki-kanaya/quegress/internal/platform/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrEmptyRedisAddr          = errors.New("redis addr is empty")
	ErrEmptyJobStateAddress    = errors.New("jobstate grpc addr is empty")
	ErrEmptyConsumerName       = errors.New("consumer name is empty")
	ErrInvalidConcurrency      = errors.New("concurrency must be positive")
	ErrInvalidRetryBackoff     = errors.New("retry backoff must be positive")
	ErrInvalidRequestTimeout   = errors.New("request timeout must be positive")
	ErrInvalidMaxResponseBytes = errors.New("max response bytes must be positive")
)

type Config struct {
	RedisAddr        string
	JobStateGRPCAddr string
	ConsumerName     string
	Concurrency      int
	RetryBackoff     time.Duration
	RequestTimeout   time.Duration
	MaxResponseBytes int64
}

func Run(ctx context.Context, cfg Config) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	rdb, err := platformredis.NewClient(ctx, platformredis.Config{
		Addr: cfg.RedisAddr,
	})
	if err != nil {
		return err
	}
	defer rdb.Close()

	queue := jobmessaginginfra.NewRedisQueue(rdb)
	if err := queue.EnsureGroup(ctx); err != nil {
		return err
	}

	conn, err := grpc.NewClient(cfg.JobStateGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	jobExecutionClient, err := executorinfra.NewGRPCJobExecutionClient(jobstatev1.NewJobStateServiceClient(conn))
	if err != nil {
		return err
	}

	httpClient, err := executorinfra.NewHTTPClient(http.DefaultClient, cfg.MaxResponseBytes)
	if err != nil {
		return err
	}

	processMessage, err := executorusecase.NewProcessMessageUseCase(queue, jobExecutionClient, httpClient, cfg.RequestTimeout)
	if err != nil {
		return err
	}

	consumerName := jobmessagingdomain.ConsumerName(cfg.ConsumerName)
	if err := consumerName.Validate(); err != nil {
		return err
	}

	log.Printf("executor started: consumer=%s concurrency=%d", consumerName, cfg.Concurrency)

	input := executorusecase.ProcessMessageInput{
		ConsumerName: consumerName,
	}

	var wg sync.WaitGroup
	wg.Add(cfg.Concurrency)
	for workerID := 1; workerID <= cfg.Concurrency; workerID++ {
		go func(workerID int) {
			defer wg.Done()

			for ctx.Err() == nil {
				if err := processMessage.Execute(ctx, input); err != nil {
					log.Printf("executor worker failed to process message: worker=%d err=%v", workerID, err)
					sleep(ctx, cfg.RetryBackoff)
				}
			}
		}(workerID)
	}

	wg.Wait()
	return nil
}

func (cfg Config) validate() error {
	if cfg.RedisAddr == "" {
		return ErrEmptyRedisAddr
	}
	if cfg.JobStateGRPCAddr == "" {
		return ErrEmptyJobStateAddress
	}
	if cfg.ConsumerName == "" {
		return ErrEmptyConsumerName
	}
	if cfg.Concurrency <= 0 {
		return ErrInvalidConcurrency
	}
	if cfg.RetryBackoff <= 0 {
		return ErrInvalidRetryBackoff
	}
	if cfg.RequestTimeout <= 0 {
		return ErrInvalidRequestTimeout
	}
	if cfg.MaxResponseBytes <= 0 {
		return ErrInvalidMaxResponseBytes
	}
	return nil
}

func sleep(ctx context.Context, duration time.Duration) {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
	case <-timer.C:
	}
}
