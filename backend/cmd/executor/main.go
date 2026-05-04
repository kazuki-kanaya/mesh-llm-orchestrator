package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/infrastructure"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/executor/usecase"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
)

func main() {
	ctx := context.Background()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb, err := redis.NewClient(ctx, redis.Config{
		Addr: redisAddr,
	})
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	repo := infrastructure.NewRedisJobRepository(rdb, 24*time.Hour)
	queue := infrastructure.NewRedisJobQueue(rdb)
	publisher := infrastructure.NewRedisJobPublisher(rdb)
	httpClient := infrastructure.NewHTTPClient(&http.Client{
		Timeout: 5 * time.Minute,
	})

	executeJobUseCase := usecase.NewExecuteJobUseCase(
		repo,
		queue,
		httpClient,
		publisher,
	)

	log.Println("executor started")

	const retryBackoff = time.Second

	for {
		if err := executeJobUseCase.Execute(ctx); err != nil {
			log.Printf("failed to execute job: %v", err)
			time.Sleep(retryBackoff)
		}
	}
}
