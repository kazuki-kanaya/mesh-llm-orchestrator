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

	redisAddr := getEnv("REDIS_ADDR")

	rdb, err := redis.NewClient(ctx, redis.Config{
		Addr: redisAddr,
	})
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer rdb.Close()

	terminalTTL := durationFromEnv("JOB_TERMINAL_TTL")
	if terminalTTL <= 0 {
		log.Fatal("JOB_TERMINAL_TTL must be positive")
	}

	repo := infrastructure.NewRedisJobRepository(rdb, terminalTTL)
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
