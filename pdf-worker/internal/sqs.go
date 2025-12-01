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
	queueURL := os.Getenv("PDF_QUEUE_URL")

	log.Println("📥 Listening:", queueURL)

	for {
		msgs, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     10,
		})

		if err != nil {
			log.Println("Receive error:", err)
			continue
		}

		if len(msgs.Messages) == 0 {
			continue
		}

		for _, m := range msgs.Messages {

			var job Job
			json.Unmarshal([]byte(*m.Body), &job)

			log.Println("⚙ Processing PDF job:", job.Tool)

			ProcessJob(job)

			sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: m.ReceiptHandle,
			})
		}
	}
}
