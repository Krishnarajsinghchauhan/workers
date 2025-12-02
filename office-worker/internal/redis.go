package internal

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var client *redis.Client

func InitRedis() {
	client = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Println("❌ Redis connection failed:", err)
	} else {
		log.Println("✅ Redis connected")
	}
}

func UpdateStatus(jobID, status string) {
	err := client.Set(ctx, "job:"+jobID, status, 0).Err()
	if err != nil {
		log.Println("❌ Redis UpdateStatus error:", err)
	}
}

func SaveResult(jobID, url string) {
	// Must be JSON array (backend requirement)
	urls := []string{url}
	b, _ := json.Marshal(urls)

	client.Set(ctx, "result:"+jobID, string(b), 0)
	client.Set(ctx, "job:"+jobID, "completed", 0)

	log.Println("💾 Saved result for job:", jobID)
}
