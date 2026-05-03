package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, cfg Config) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
