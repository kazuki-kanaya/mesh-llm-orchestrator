package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, app.Config{
		RedisAddr:        mustGetEnv("REDIS_ADDR"),
		JobStateGRPCAddr: mustGetEnv("JOBSTATE_GRPC_ADDR"),
		ConsumerName:     mustGetEnv("EXECUTOR_CONSUMER_NAME"),
		RetryBackoff:     mustDurationEnv("EXECUTOR_RETRY_BACKOFF"),
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
