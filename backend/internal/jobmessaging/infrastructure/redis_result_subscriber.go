package infrastructure

import (
	"context"
	"errors"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobmessaging/ports"
	jobstatedomain "github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/jobstate/domain"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

var (
	ErrNilRedisClient              = errors.New("redis client is nil")
	ErrJobResultSubscriptionClosed = errors.New("job result subscription closed")
)

type RedisResultSubscriber struct {
	rdb *goredis.Client
}

var _ ports.JobResultSubscriber = (*RedisResultSubscriber)(nil)
var _ ports.JobResultSubscription = (*redisResultSubscription)(nil)

func NewRedisResultSubscriber(rdb *goredis.Client) (ports.JobResultSubscriber, error) {
	if rdb == nil {
		return nil, ErrNilRedisClient
	}

	return &RedisResultSubscriber{
		rdb: rdb,
	}, nil
}

func (s *RedisResultSubscriber) Subscribe(ctx context.Context, jobID jobstatedomain.JobID) (ports.JobResultSubscription, error) {
	if err := jobID.Validate(); err != nil {
		return nil, err
	}

	pubsub := s.rdb.Subscribe(ctx, redis.JobResultChannel(jobID.String()))
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, err
	}

	return &redisResultSubscription{
		pubsub: pubsub,
		ch:     pubsub.Channel(),
	}, nil
}

type redisResultSubscription struct {
	pubsub *goredis.PubSub
	ch     <-chan *goredis.Message
}

func (s *redisResultSubscription) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case _, ok := <-s.ch:
		if !ok {
			return ErrJobResultSubscriptionClosed
		}
		return nil
	}
}

func (s *redisResultSubscription) Close() error {
	return s.pubsub.Close()
}
