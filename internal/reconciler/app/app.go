package app

import (
	"context"
	"errors"
	"log"
	"time"

	jobstatev1 "github.com/kazuki-kanaya/quegress/gen/go/jobstate/v1"
	jobmessagingdomain "github.com/kazuki-kanaya/quegress/internal/jobmessaging/domain"
	jobmessaginginfra "github.com/kazuki-kanaya/quegress/internal/jobmessaging/infrastructure"
	platformredis "github.com/kazuki-kanaya/quegress/internal/platform/redis"
	reconcilerinfra "github.com/kazuki-kanaya/quegress/internal/reconciler/infrastructure"
	reconcilerusecase "github.com/kazuki-kanaya/quegress/internal/reconciler/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrEmptyRedisAddr       = errors.New("redis addr is empty")
	ErrEmptyJobStateAddress = errors.New("jobstate grpc addr is empty")
	ErrEmptyConsumerName    = errors.New("consumer name is empty")
	ErrInvalidStaleAfter    = errors.New("stale after must be positive")
	ErrInvalidBatchSize     = errors.New("batch size must be positive")
	ErrInvalidInterval      = errors.New("interval must be positive")
)

type Config struct {
	RedisAddr        string
	JobStateGRPCAddr string
	ConsumerName     string
	StaleAfter       time.Duration
	BatchSize        int64
	Interval         time.Duration
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

	client, err := reconcilerinfra.NewGRPCJobReconcileClient(jobstatev1.NewJobStateServiceClient(conn))
	if err != nil {
		return err
	}

	reconcile, err := reconcilerusecase.NewReconcileStalePendingJobsUseCase(
		queue,
		client,
		cfg.StaleAfter,
		cfg.BatchSize,
	)
	if err != nil {
		return err
	}

	consumerName := jobmessagingdomain.ConsumerName(cfg.ConsumerName)
	if err := consumerName.Validate(); err != nil {
		return err
	}

	input := reconcilerusecase.ReconcileStalePendingJobsInput{
		ConsumerName: consumerName,
	}
	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	log.Printf("reconciler started: consumer=%s interval=%s", consumerName, cfg.Interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			output, err := reconcile.Execute(ctx, input)
			if err != nil {
				log.Printf("failed to reconcile stale pending jobs: %v", err)
				continue
			}
			logOutput(output)
		}
	}
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
	if cfg.StaleAfter <= 0 {
		return ErrInvalidStaleAfter
	}
	if cfg.BatchSize <= 0 {
		return ErrInvalidBatchSize
	}
	if cfg.Interval <= 0 {
		return ErrInvalidInterval
	}
	return nil
}

func logOutput(output *reconcilerusecase.ReconcileStalePendingJobsOutput) {
	if output == nil {
		return
	}
	if output.Recovered == 0 && output.AckedTerminal == 0 && output.AckedQueued == 0 {
		return
	}

	log.Printf(
		"reconciled stale pending jobs: claimed=%d recovered=%d acked_terminal=%d acked_queued=%d",
		output.Claimed,
		output.Recovered,
		output.AckedTerminal,
		output.AckedQueued,
	)
}
