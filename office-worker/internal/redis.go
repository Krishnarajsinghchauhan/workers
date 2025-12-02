package internal

import (
	"context"
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
	key := "job:" + jobID
	err := client.Set(ctx, key, status, 0).Err()
	if err != nil {
		log.Println("❌ Redis UpdateStatus error:", err)
	}
}

func SaveResult(jobID, url string) {
	// Save result URL
	client.Set(ctx, "result:"+jobID, url, 0)

	// Mark job completed
	client.Set(ctx, "job:"+jobID, "completed", 0)
}
