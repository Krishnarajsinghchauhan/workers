package main

import (
    "log"
    "os"
    "pdf-worker/internal"

    "github.com/joho/godotenv"
)

func main() {
    log.Println("PDF Worker Started")

    // Try to load .env from multiple locations
    paths := []string{
        ".env",
        "../.env",
        "../../.env",
        "/Users/krishna/Personal Project/pdf/workers/pdf-worker/.env",
    }

    for _, p := range paths {
        if err := godotenv.Load(p); err == nil {
            log.Println("Loaded .env from:", p)
            break
        }
    }

    // Verify env loaded
    log.Println("AWS_S3_BUCKET =", os.Getenv("AWS_S3_BUCKET"))
    log.Println("PDF_QUEUE_URL =", os.Getenv("PDF_QUEUE_URL"))

    internal.InitSQS()
    internal.InitRedis()
    internal.InitS3()

    internal.ListenToQueue()
}
