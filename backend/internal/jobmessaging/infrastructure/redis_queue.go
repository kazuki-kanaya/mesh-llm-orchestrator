package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/ports"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	platformredis "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	"github.com/redis/go-redis/v9"
)

const (
	defaultReadCount = int64(1)
	defaultReadBlock = 5 * time.Second
	xAutoClaimStart  = "0-0"
)

type RedisQueue struct {
	rdb *redis.Client
}

func NewRedisQueue(rdb *redis.Client) ports.JobQueue {
	return &RedisQueue{
		rdb: rdb,
	}
}

func (q *RedisQueue) EnsureGroup(ctx context.Context) error {
	err := q.rdb.XGroupCreateMkStream(
		ctx,
		platformredis.JobStreamKey(),
		platformredis.JobConsumerGroupName(),
		"0",
	).Err()
	if err == nil {
		return nil
	}
	if strings.HasPrefix(err.Error(), "BUSYGROUP ") {
		return nil
	}
	return err
}

func (q *RedisQueue) Read(ctx context.Context, consumerName domain.ConsumerName) (*domain.Message, error) {
	if err := consumerName.Validate(); err != nil {
		return nil, err
	}

	streams, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    platformredis.JobConsumerGroupName(),
		Consumer: consumerName.String(),
		Streams:  []string{platformredis.JobStreamKey(), ">"},
		Count:    defaultReadCount,
		Block:    defaultReadBlock,
	}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, nil
	}

	return messageFromRedis(streams[0].Messages[0])
}

func (q *RedisQueue) ClaimStalePending(ctx context.Context, consumerName domain.ConsumerName, minIdle time.Duration, count int64) ([]*domain.Message, error) {
	if err := consumerName.Validate(); err != nil {
		return nil, err
	}
	if minIdle <= 0 {
		return nil, fmt.Errorf("min idle must be positive: %s", minIdle)
	}
	if count <= 0 {
		return nil, fmt.Errorf("count must be positive: %d", count)
	}

	// Start from the beginning for now. If the pending list grows large,
	// expose the XAUTOCLAIM cursor to avoid repeatedly scanning from "0-0".
	messages, _, err := q.rdb.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Stream:   platformredis.JobStreamKey(),
		Group:    platformredis.JobConsumerGroupName(),
		Consumer: consumerName.String(),
		MinIdle:  minIdle,
		Start:    xAutoClaimStart,
		Count:    count,
	}).Result()
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Message, 0, len(messages))
	for _, message := range messages {
		msg, err := messageFromRedis(message)
		if err != nil {
			return nil, err
		}
		result = append(result, msg)
	}

	return result, nil
}

func (q *RedisQueue) Ack(ctx context.Context, messageID domain.MessageID) error {
	if err := messageID.Validate(); err != nil {
		return err
	}

	return q.rdb.XAck(
		ctx,
		platformredis.JobStreamKey(),
		platformredis.JobConsumerGroupName(),
		messageID.String(),
	).Err()
}

func messageFromRedis(message redis.XMessage) (*domain.Message, error) {
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
