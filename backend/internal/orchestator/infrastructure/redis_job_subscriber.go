package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestator/ports"
	"github.com/redis/go-redis/v9"
)

type RedisJobSubscriber struct {
	rdb *redis.Client
}

func NewRedisJobSubscriber(rdb *redis.Client) *RedisJobSubscriber {
	return &RedisJobSubscriber{
		rdb: rdb,
	}
}

func (s *RedisJobSubscriber) Subscribe(ctx context.Context, jobID uuid.UUID) (ports.Subscription, error) {
	sub := s.rdb.Subscribe(ctx, jobResultChannel(jobID))
	if _, err := sub.Receive(ctx); err != nil {
		_ = sub.Close()
		return nil, err
	}

	ch := make(chan struct{}, 1)

	go func() {
		defer close(ch)

		select {
		case _, ok := <-sub.Channel():
			if !ok {
				return
			}
			ch <- struct{}{}

		case <-ctx.Done():
			return
		}
	}()

	return &RedisJobSubscription{
		sub: sub,
		ch:  ch,
	}, nil
}

type RedisJobSubscription struct {
	sub *redis.PubSub
	ch  chan struct{}
}

func (s *RedisJobSubscription) Channel() <-chan struct{} {
	return s.ch
}

func (s *RedisJobSubscription) Close() error {
	return s.sub.Close()
}

func jobResultChannel(jobID uuid.UUID) string {
	return "result:" + jobID.String()
}

var _ ports.JobSubscriber = (*RedisJobSubscriber)(nil)
var _ ports.Subscription = (*RedisJobSubscription)(nil)
