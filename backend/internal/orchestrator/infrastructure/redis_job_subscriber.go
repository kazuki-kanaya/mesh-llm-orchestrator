package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/ports"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	goredis "github.com/redis/go-redis/v9"
)

type RedisJobSubscriber struct {
	rdb *goredis.Client
}

func NewRedisJobSubscriber(rdb *goredis.Client) *RedisJobSubscriber {
	return &RedisJobSubscriber{
		rdb: rdb,
	}
}

func (s *RedisJobSubscriber) Subscribe(ctx context.Context, jobID uuid.UUID) (ports.Subscription, error) {
	sub := s.rdb.Subscribe(ctx, redis.JobResultChannel(jobID))
	if _, err := sub.Receive(ctx); err != nil {
		_ = sub.Close()
		return nil, err
	}

	ch := make(chan struct{}, 1)

	go func() {
		defer close(ch)

		for {
			select {
			case _, ok := <-sub.Channel():
				if !ok {
					return
				}

				select {
				case ch <- struct{}{}:
				default:
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return &RedisJobSubscription{
		sub: sub,
		ch:  ch,
	}, nil
}

type RedisJobSubscription struct {
	sub *goredis.PubSub
	ch  chan struct{}
}

func (s *RedisJobSubscription) Channel() <-chan struct{} {
	return s.ch
}

func (s *RedisJobSubscription) Close() error {
	return s.sub.Close()
}

var _ ports.JobSubscriber = (*RedisJobSubscriber)(nil)
var _ ports.Subscription = (*RedisJobSubscription)(nil)
