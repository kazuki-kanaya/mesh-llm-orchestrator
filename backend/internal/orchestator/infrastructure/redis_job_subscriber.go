package infrastructure

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisJobSubscriber struct {
	rdb *redis.Client
}

func (s *RedisJobSubscriber) Subscribe(ctx context.Context, jobID uuid.UUID) *RedisJobSubscription {
	sub := s.rdb.Subscribe(ctx, jobResultChannel(jobID))
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
	}
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
