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

	store := infrastructure.NewRedisJobRecoveryStore(rdb)
	recoverStaleJobsUseCase := usecase.NewRecoverStaleJobsUseCase(
		store,
		durationFromEnv("RECOVERY_STALE_AFTER", 5*time.Minute),
		int64FromEnv("RECOVERY_BATCH_SIZE", 100),
	)

	interval := durationFromEnv("RECOVERY_INTERVAL", time.Minute)
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

func durationFromEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("invalid duration env: key=%s value=%q err=%v", key, value, err)
		return fallback
	}

	return duration
}

func int64FromEnv(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		log.Printf("invalid int env: key=%s value=%q err=%v", key, value, err)
		return fallback
	}

	return parsed
}
