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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	targetBaseURL := os.Getenv("TARGET_BASE_URL")
	if targetBaseURL == "" {
		log.Fatal("TARGET_BASE_URL is required")
	}

	addr := os.Getenv("ORCHESTRATOR_ADDR")
	if addr == "" {
		addr = ":8080"
	}

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
		30*time.Second,
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
