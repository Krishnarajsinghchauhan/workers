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

// ----------------------------
// Initialize S3
// ----------------------------
func InitS3() {
	bucket = os.Getenv("AWS_S3_BUCKET")
	if bucket == "" {
		log.Println("❌ AWS_S3_BUCKET is EMPTY!")
	} else {
		log.Println("📦 Using S3 Bucket:", bucket)
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-1"), // Bucket region (IMPORTANT)
	)
	if err != nil {
		log.Println("❌ S3 config error:", err)
		return
	}

	s3Client = s3.NewFromConfig(cfg)
	log.Println("✅ S3 initialized (Region: us-east-1)")
}

// ----------------------------
// Extract S3 Key
// Supports:
//   s3://bucket/key
//   https://bucket.s3.amazonaws.com/key
// ----------------------------
func ExtractS3Key(fileURL string) string {

	// Format 1 → s3://bucket/key
	if strings.HasPrefix(fileURL, "s3://") {
		trim := strings.TrimPrefix(fileURL, "s3://")
		parts := strings.SplitN(trim, "/", 2)

		if len(parts) < 2 {
			log.Println("❌ Missing key in S3 URL:", fileURL)
			return ""
		}

		return parts[1]
	}

	// Format 2 → https://bucket.s3.amazonaws.com/key
	httpsPrefix := "https://" + bucket + ".s3.amazonaws.com/"
	if strings.HasPrefix(fileURL, httpsPrefix) {
		return fileURL[len(httpsPrefix):]
	}

	log.Println("❌ Invalid S3 URL:", fileURL)
	return ""
}

// ----------------------------
// Download file FROM S3
// ----------------------------
func DownloadFromS3(fileURL string) string {
	key := ExtractS3Key(fileURL)

	if key == "" {
		log.Println("❌ Cannot extract S3 key from URL:", fileURL)
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

	log.Println("⬇ Downloaded:", key, "→", localPath)
	return localPath
}

// ----------------------------
// Upload file TO S3
// ALWAYS publicly readable
// ----------------------------
func UploadToS3(localPath string) string {
	filename := filepath.Base(localPath)
	key := "processed/" + filename

	file, _ := os.Open(localPath)
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
