package internal

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client
var bucket string

func InitS3() {
	bucket = os.Getenv("AWS_S3_BUCKET")

	if bucket == "" {
		log.Println("❌ AWS_S3_BUCKET is EMPTY")
	} else {
		log.Println("📦 Using S3 Bucket:", bucket)
	}

	cfg, _ := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"),
	)

	s3Client = s3.NewFromConfig(cfg)
	log.Println("✅ S3 initialized (Region: us-east-1)")
}

func DownloadFromS3(url string) string {
	key := ExtractS3Key(url)
	if key == "" {
		return ""
	}

	local := filepath.Join("/tmp", filepath.Base(key))

	out, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		log.Println("❌ S3 download failed:", err)
		return ""
	}

	file, _ := os.Create(local)
	io.Copy(file, out.Body)
	file.Close()

	log.Println("⬇ Downloaded:", key, "→", local)
	return local
}

func UploadToS3(local string) string {
	key := "processed/" + filepath.Base(local)

	file, _ := os.Open(local)
	defer file.Close()

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   file,
	})
	if err != nil {
		log.Println("❌ Upload failed:", err)
		return ""
	}

	url := "https://" + bucket + ".s3.amazonaws.com/" + key
	log.Println("⬆ Uploaded:", url)

	return url
}
