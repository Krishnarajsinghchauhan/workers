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
	ID      string            `json:"id"`
	Tool    string            `json:"tool"`
	Files   []string          `json:"files"`
	Options map[string]string `json:"options"`
}

func InitSQS() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	sqsClient = sqs.NewFromConfig(cfg)
	log.Println("✅ SQS initialized")
}

func ListenToQueue() {
	queueURL := os.Getenv("OCR_QUEUE_URL")
	if queueURL == "" {
		log.Println("❌ OCR_QUEUE_URL is EMPTY")
		return
	}

	log.Println("📥 Listening to:", queueURL)

	for {
		msgs, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     10,
		})
		if err != nil {
			log.Println("❌ Receive error:", err)
			continue
		}

		if len(msgs.Messages) == 0 {
			continue
		}

		for _, m := range msgs.Messages {
			var job Job
			json.Unmarshal([]byte(*m.Body), &job)

			ProcessJob(job)

			sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: m.ReceiptHandle,
			})
		}
	}
}
