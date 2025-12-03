package main

import (
	"log"
	"os"
	"image-worker/internal"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Image Worker Started")

	// Load .env from possible locations
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
		log.Println("⚠️ WARNING: No .env file loaded!")
	}

	// print ENV
	log.Println("IMAGE_QUEUE_URL =", os.Getenv("IMAGE_QUEUE_URL"))
	log.Println("AWS_S3_BUCKET =", os.Getenv("AWS_S3_BUCKET"))

	internal.InitS3()
	internal.InitRedis()
	internal.InitSQS()

	internal.ListenToQueue()
}
