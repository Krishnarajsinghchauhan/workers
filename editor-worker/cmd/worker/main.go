package main

import (
	"log"
	"os"
	"editor-worker/internal"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Editor Worker Started")

	// Load .env like other workers
	paths := []string{
		".env",
		"../.env",
		"../../.env",
		"/Users/krishna/Personal Project/pdf/workers/image-worker/.env",
	}

	loaded := false
	for _, p := range paths {
		if err := godotenv.Load(p); err == nil {
			log.Println("Loaded .env from:", p)
			loaded = true
			break
		}
	}

	if !loaded {
		log.Println("⚠ WARNING: .env not found. Using system env.")
	}

	log.Println("EDITOR_QUEUE_URL =", os.Getenv("EDITOR_QUEUE_URL"))
	log.Println("AWS_S3_BUCKET =", os.Getenv("AWS_S3_BUCKET"))
	log.Println("REDIS_HOST =", os.Getenv("REDIS_HOST"))

	internal.InitSQS()
	internal.InitS3()
	internal.InitRedis()

	internal.ListenToQueue()
}
