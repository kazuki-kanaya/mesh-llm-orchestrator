package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/kazuki-kanaya/quegress/internal/executor/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, app.Config{
		RedisAddr:        mustGetEnv("REDIS_ADDR"),
		JobStateGRPCAddr: mustGetEnv("JOBSTATE_GRPC_ADDR"),
		ConsumerName:     mustGetEnv("EXECUTOR_CONSUMER_NAME"),
		Concurrency:      mustIntEnv("EXECUTOR_CONCURRENCY"),
		RetryBackoff:     mustDurationEnv("EXECUTOR_RETRY_BACKOFF"),
		RequestTimeout:   mustDurationEnv("EXECUTOR_REQUEST_TIMEOUT"),
		MaxResponseBytes: mustInt64Env("EXECUTOR_MAX_RESPONSE_BYTES"),
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
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n <= 0 {
		log.Fatalf("%s must be a positive integer: %q", key, value)
	}
	return n
}

func mustIntEnv(key string) int {
	value := mustGetEnv(key)
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		log.Fatalf("%s must be a positive integer: %q", key, value)
	}
	return n
}
