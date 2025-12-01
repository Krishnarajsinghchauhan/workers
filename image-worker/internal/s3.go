package internal

import (
	"context"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client
var bucket string

func InitS3() {
	bucket = os.Getenv("AWS_S3_BUCKET")
	if bucket == "" {
		log.Println("❌ AWS_S3_BUCKET is EMPTY")
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Println("❌ S3 config error:", err)
		return
	}

	s3Client = s3.NewFromConfig(cfg)
	log.Println("✅ S3 initialized (Region: us-east-1)")
}

// Extract key
func ExtractS3Key(url string) string {

	// Format 1: s3://bucket/key
	if strings.HasPrefix(url, "s3://") {
		trim := strings.TrimPrefix(url, "s3://")
		parts := strings.SplitN(trim, "/", 2)
		if len(parts) < 2 {
			return ""
		}
		return parts[1]
	}

	// Format 2: https://bucket.s3.amazonaws.com/key
	prefix := "https://" + bucket + ".s3.amazonaws.com/"
	if strings.HasPrefix(url, prefix) {
		return strings.TrimPrefix(url, prefix)
	}

	log.Println("❌ Invalid S3 URL:", url)
	return ""
}

// Download file
func DownloadFromS3(url string) string {
	key := ExtractS3Key(url)
	if key == "" {
		log.Println("❌ Cannot extract key from:", url)
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

// Upload file
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
		log.Println("❌ S3 upload failed:", err)
		return ""
	}

	url := "https://" + bucket + ".s3.amazonaws.com/" + key
	log.Println("⬆ Uploaded:", url)
	return url
}
