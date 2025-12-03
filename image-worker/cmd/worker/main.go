package main

import (
    "log"
    "os"
    "ocr-worker/internal"

    "github.com/joho/godotenv"
)

func main() {
    log.Println("OCR Worker Started")

    paths := []string{
        ".env",
        "../.env",
        "../../.env",
        "/Users/krishna/Personal Project/pdf/workers/ocr-worker/.env",
    }

    for _, p := range paths {
        if err := godotenv.Load(p); err == nil {
            log.Println("Loaded .env from:", p)
            break
        }
    }

    log.Println("OCR_QUEUE_URL =", os.Getenv("OCR_QUEUE_URL"))
    log.Println("AWS_S3_BUCKET =", os.Getenv("AWS_S3_BUCKET"))

    internal.InitRedis()
    internal.InitSQS()
    internal.InitS3()

    internal.ListenToQueue()
}
