package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobqueue/ports"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

const (
	defaultReadCount = int64(1)
	defaultReadBlock = 5 * time.Second
)

type RedisQueue struct {
	rdb   *goredis.Client
	count int64
	block time.Duration
}

func NewRedisQueue(rdb *goredis.Client) ports.JobQueue {
	return &RedisQueue{
		rdb:   rdb,
		count: defaultReadCount,
		block: defaultReadBlock,
	}
}

func (q *RedisQueue) EnsureGroup(ctx context.Context) error {
	err := q.rdb.XGroupCreateMkStream(
		ctx,
		redis.JobStreamKey(),
		redis.JobConsumerGroupName(),
		"0",
	).Err()
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func (q *RedisQueue) Read(ctx context.Context, consumerName domain.ConsumerName) (*domain.Message, error) {
	if err := consumerName.Validate(); err != nil {
		return nil, err
	}

	streams, err := q.rdb.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    redis.JobConsumerGroupName(),
		Consumer: consumerName.String(),
		Streams:  []string{redis.JobStreamKey(), ">"},
		Count:    defaultReadCount,
		Block:    defaultReadBlock,
	}).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, nil
	}

	return messageFromRedis(streams[0].Messages[0])
}

func (q *RedisQueue) Ack(ctx context.Context, messageID domain.MessageID) error {
	if err := messageID.Validate(); err != nil {
		return err
	}

	return q.rdb.XAck(
		ctx,
		redis.JobStreamKey(),
		redis.JobConsumerGroupName(),
		messageID.String(),
	).Err()
}

func messageFromRedis(message goredis.XMessage) (*domain.Message, error) {
	rawJobID, ok := message.Values["job_id"]
	if !ok {
		return nil, fmt.Errorf("%w: missing job_id", domain.ErrInvalidMessage)
	}

	jobIDValue, ok := rawJobID.(string)
	if !ok || jobIDValue == "" {
		return nil, fmt.Errorf("%w: invalid job_id", domain.ErrInvalidMessage)
	}

	jobID, err := jobstatedomain.ParseJobID(jobIDValue)
	if err != nil {
		return nil, err
	}

	msg := &domain.Message{
		ID:    domain.MessageID(message.ID),
		JobID: jobID,
	}
	if err := msg.ID.Validate(); err != nil {
		return nil, err
	}

	return msg, nil
}
