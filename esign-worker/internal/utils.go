package internal

import (
	"os"
	"path/filepath"
	"time"
	"log"
	"strings"
)

// ------------------------------------------------------------
// TEMP FILE HELPERS
// ------------------------------------------------------------
func TempName(prefix, ext string) string {
	return filepath.Join("/tmp", prefix+"_"+time.Now().Format("150405")+ext)
}

func DeleteFile(path string) {
	os.Remove(path)
}

// ------------------------------------------------------------
// S3 KEY EXTRACTOR
// Supports:
// 1) s3://bucket/key
// 2) https://bucket.s3.amazonaws.com/key
// ------------------------------------------------------------
func ExtractS3Key(raw string) string {
	bucket := os.Getenv("AWS_S3_BUCKET")

	// Format 1
	if strings.HasPrefix(raw, "s3://") {
		trim := strings.TrimPrefix(raw, "s3://")
		parts := strings.SplitN(trim, "/", 2)
		if len(parts) < 2 {
			log.Println("❌ Missing key in S3 URL:", raw)
			return ""
		}
		return parts[1]
	}

	// Format 2
	prefix := "https://" + bucket + ".s3.amazonaws.com/"
	if strings.HasPrefix(raw, prefix) {
		return raw[len(prefix):]
	}

	log.Println("❌ Invalid S3 URL (cannot extract key):", raw)
	return ""
}
