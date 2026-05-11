package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kazuki-kanaya/quegress/internal/reconciler/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, app.Config{
		RedisAddr:        mustGetEnv("REDIS_ADDR"),
		JobStateGRPCAddr: mustGetEnv("JOBSTATE_GRPC_ADDR"),
		ConsumerName:     mustGetEnv("RECONCILER_CONSUMER_NAME"),
		StaleAfter:       mustDurationEnv("RECONCILER_STALE_AFTER"),
		BatchSize:        mustInt64Env("RECONCILER_BATCH_SIZE"),
		Interval:         mustDurationEnv("RECONCILER_INTERVAL"),
	}); err != nil {
		log.Fatal(err)
	}
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func mustDurationEnv(key string) time.Duration {
	value := mustGetEnv(key)
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		log.Fatalf("%s must be a positive duration: %q", key, value)
	}
	return duration
}

func mustInt64Env(key string) int64 {
	value := mustGetEnv(key)
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil || result <= 0 {
		log.Fatalf("%s must be a positive integer: %q", key, value)
	}
	return result
}
