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
	Options map[string]string `json:"options"` // watermark text, position, etc.
}

func InitSQS() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	sqsClient = sqs.NewFromConfig(cfg)
}

func ListenToQueue() {
	queueURL := os.Getenv("EDITOR_QUEUE_URL")

	for {
		msgs, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     20,
		})

		for _, msg := range msgs.Messages {
			var job Job
			json.Unmarshal([]byte(*msg.Body), &job)

			log.Println("Editor job received:", job.Tool)

			ProcessJob(job)

			sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: msg.ReceiptHandle,
			})
		}
	}
}
