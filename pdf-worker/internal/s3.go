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

func ExtractS3Key(fileURL string) string {

	if strings.HasPrefix(fileURL, "s3://") {
		trim := strings.TrimPrefix(fileURL, "s3://")
		parts := strings.SplitN(trim, "/", 2)
		if len(parts) < 2 {
			return ""
		}
		return parts[1]
	}

	prefix := "https://" + bucket + ".s3.amazonaws.com/"
	if strings.HasPrefix(fileURL, prefix) {
		return fileURL[len(prefix):]
	}

	log.Println("❌ Invalid S3 URL:", fileURL)
	return ""
}

func DownloadFromS3(url string) string {
	key := ExtractS3Key(url)
	if key == "" {
		return ""
	}

	localPath := filepath.Join("/tmp", filepath.Base(key))

	out, err := s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		log.Println("❌ S3 download failed:", err)
		return ""
	}

	file, _ := os.Create(localPath)
	io.Copy(file, out.Body)
	file.Close()

	log.Println("⬇ Downloaded:", key)
	return localPath
}

func UploadToS3(path string) string {
	filename := filepath.Base(path)
	key := "processed/" + filename

	f, _ := os.Open(path)
	defer f.Close()

	_, err := s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})

	if err != nil {
		log.Println("❌ Upload failed:", err)
		return ""
	}

	url := "https://" + bucket + ".s3.amazonaws.com/" + key
	log.Println("⬆ Uploaded:", url)

	return url
}
