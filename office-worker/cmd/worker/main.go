package main

import (
	"log"
	"os"
	"office-worker/internal"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Office Worker Started")

	paths := []string{
		".env",
		"../.env",
		"../../.env",
		`/Users/krishna/Personal Project/pdf/workers/office-worker/.env`,
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
		log.Println("⚠️ No .env file loaded!")
	}

	log.Println("OFFICE_QUEUE_URL =", os.Getenv("OFFICE_QUEUE_URL"))

	internal.InitRedis()
	internal.InitSQS()
	internal.InitS3()

	internal.ListenToQueue()
}
