package internal

import (
	"context"

	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var redisClient *redis.Client

func InitRedis() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
}

func UpdateStatus(id, status string) {
	redisClient.Set(ctx, "job:"+id, status, 0)
}

func SaveResult(id, url string) {
	redisClient.Set(ctx, "result:"+id, url, 0)
}
