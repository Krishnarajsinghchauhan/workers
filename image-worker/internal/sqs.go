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
	cfg, _ := config.LoadDefaultConfig(context.TODO())
	sqsClient = sqs.NewFromConfig(cfg)
}

type Job struct {
	ID    string   `json:"id"`
	Tool  string   `json:"tool"`
	Files []string `json:"files"`
}

func ListenToQueue() {
	queueURL := os.Getenv("IMAGE_QUEUE_URL")

	for {
		msg, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &queueURL,
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     20,
		})

		for _, m := range msg.Messages {
			var job Job
			json.Unmarshal([]byte(*m.Body), &job)

			log.Println("Image job received:", job.Tool)

			ProcessJob(job)

			sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
				QueueUrl:      &queueURL,
				ReceiptHandle: m.ReceiptHandle,
			})
		}
	}
}
