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
    log.Println("✅ SQS initialized")
}

func ListenToQueue() {
    queueURL := os.Getenv("OCR_QUEUE_URL")

    log.Println("📥 Listening to:", queueURL)

    for {
        msgs, _ := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
            QueueUrl:            &queueURL,
            MaxNumberOfMessages: 1,
            WaitTimeSeconds:     20,
        })

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
