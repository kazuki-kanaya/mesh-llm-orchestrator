package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/platform/redis"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/recovery/infrastructure"
	"github.com/kazuki-kanaya/mesh-llm-orchestrator/backend/internal/recovery/usecase"
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

	store := infrastructure.NewRedisJobRecoveryStore(rdb)
	recoverStaleJobsUseCase := usecase.NewRecoverStaleJobsUseCase(
		store,
		durationFromEnv("RECOVERY_STALE_AFTER"),
		int64FromEnv("RECOVERY_BATCH_SIZE"),
	)

	interval := durationFromEnv("RECOVERY_INTERVAL")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("recovery started: interval=%s", interval)

	for {
		recovered, err := recoverStaleJobsUseCase.Execute(ctx, time.Now().UTC())
		if err != nil {
			log.Printf("failed to recover stale jobs: %v", err)
		}
		if recovered > 0 {
			log.Printf("recovered stale jobs: count=%d", recovered)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
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

func int64FromEnv(key string) int64 {
	value := getEnv(key)
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Fatalf("invalid int env: key=%s value=%q err=%v", key, value, err)
	}

	return parsed
}
