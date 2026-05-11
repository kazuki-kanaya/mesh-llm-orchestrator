package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kazuki-kanaya/quegress/internal/jobstate/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, app.Config{
		RedisAddr: mustGetEnv("REDIS_ADDR"),
		GRPCAddr:  mustGetEnv("JOBSTATE_GRPC_ADDR"),
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
