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

// Init
func InitS3() {
	bucket = os.Getenv("AWS_S3_BUCKET")

	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	s3Client = s3.NewFromConfig(cfg)
}

// Extract S3 key
func ExtractS3Key(url string) string {

	prefix := "https://" + bucket + ".s3.amazonaws.com/"
	if strings.HasPrefix(url, prefix) {
		return url[len(prefix):]
	}

	return ""
}

// Download
func DownloadFromS3(url string) string {
	key := ExtractS3Key(url)

	local := filepath.Join("/tmp", filepath.Base(key))

	resp, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		log.Println("❌ Download failed:", err)
		return ""
	}

	f, _ := os.Create(local)
	io.Copy(f, resp.Body)
	f.Close()

	return local
}

// Upload
func UploadToS3(local string) string {

	f, _ := os.Open(local)
	defer f.Close()

	key := "processed/" + filepath.Base(local)

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})

	if err != nil {
		log.Println("❌ Upload failed:", err)
		return ""
	}

	return "https://" + bucket + ".s3.amazonaws.com/" + key
}
