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

// Save job status exactly how backend expects
func UpdateStatus(jobID, status string) {
	err := client.Set(ctx, "job:"+jobID, status, 0).Err()
	if err != nil {
		log.Println("❌ Redis UpdateStatus error:", err)
	}
}

// Save result exactly how backend expects
func SaveResult(jobID, url string) {

	// Save final result URL
	client.Set(ctx, "result:"+jobID, url, 0)

	// Mark job completed
	client.Set(ctx, "job:"+jobID, "completed", 0)
}
