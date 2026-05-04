package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/infrastructure"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/presentation"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/orchestrator/usecase"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
)

func main() {
	ctx := context.Background()

	redisAddr := getEnv("REDIS_ADDR")
	targetBaseURL := getEnv("TARGET_BASE_URL")
	addr := getEnv("ORCHESTRATOR_ADDR")

	rdb, err := redis.NewClient(ctx, redis.Config{
		Addr: redisAddr,
	})
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	reader := infrastructure.NewRedisJobReader(rdb)
	creator := infrastructure.NewRedisJobCreator(rdb)
	subscriber := infrastructure.NewRedisJobSubscriber(rdb)

	createJobUseCase := usecase.NewCreateJobUseCase(creator)
	waitJobUseCase := usecase.NewWaitJobUseCase(reader, subscriber)

	handler := presentation.NewProxyHandler(
		createJobUseCase,
		waitJobUseCase,
		targetBaseURL,
		durationFromEnv("ORCHESTRATOR_WAIT_TIMEOUT"),
	)

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("orchestrator listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("orchestrator server stopped: %v", err)
	}
}

func getEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func durationFromEnv(key string) time.Duration {
	value := getEnv(key)
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("invalid duration env: key=%s value=%q err=%v", key, value, err)
	}

	return duration
}
