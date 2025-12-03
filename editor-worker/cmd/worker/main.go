package main

import (
	"log"
	"os"
	"editor-worker/internal"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Editor Worker Started")

	// Load .env from multiple paths
	paths := []string{
		".env",
		"../.env",
		"../../.env",
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
		log.Println("⚠ Using system environment variables (no .env found)")
	}

	log.Println("EDITOR_QUEUE_URL =", os.Getenv("EDITOR_QUEUE_URL"))
	log.Println("AWS_S3_BUCKET =", os.Getenv("AWS_S3_BUCKET"))

	internal.InitRedis()
	internal.InitS3()
	internal.InitSQS()

	internal.ListenToQueue()
}
