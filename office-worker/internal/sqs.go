package internal

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var sqsClient *sqs.Client

type Job struct {
	ID    string   `json:"id"`
	Tool  string   `json:"tool"`
	Files []string `json:"files"`
}

func InitSQS() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println("❌ SQS config error:", err)
		return
	}

	sqsClient = sqs.NewFromConfig(cfg)
	log.Println("✅ SQS initialized")
}

func ListenToQueue() {
	queueURL := os.Getenv("OFFICE_QUEUE_URL")

	log.Println("Loaded OFFICE_QUEUE_URL =", queueURL) // 🔹 ADD THIS

	log.Println("📥 Listening to queue:", queueURL)

	for {
		resp, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     20,
		})

		if err != nil {
			log.Println("❌ SQS receive error:", err)
			continue
		}

		for _, m := range resp.Messages {

			var job Job
			json.Unmarshal([]byte(*m.Body), &job)

			log.Println("📦 Office job received:", job.Tool)

			ProcessJob(job)

			sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: m.ReceiptHandle,
			})
		}
	}
}
