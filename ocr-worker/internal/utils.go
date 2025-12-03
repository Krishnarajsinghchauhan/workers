
package internal

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var asciiCleaner = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// SAFE FILENAME
func SafeFilename(path string) string {
	base := filepath.Base(path)
	clean := asciiCleaner.ReplaceAllString(base, "_")
	newPath := "/tmp/" + clean

	err := os.Rename(path, newPath)
	if err != nil {
		log.Println("⚠️ rename failed:", err)
		return path
	}

	return newPath
}

// EXTRACT S3 KEY
func ExtractS3Key(raw string) string {
	bucket := os.Getenv("AWS_S3_BUCKET")

	if strings.HasPrefix(raw, "s3://") {
		trim := strings.TrimPrefix(raw, "s3://")
		parts := strings.SplitN(trim, "/", 2)
		if len(parts) < 2 {
			log.Println("❌ Missing key:", raw)
			return ""
		}
		return parts[1]
	}

	prefix := "https://" + bucket + ".s3.amazonaws.com/"
	if strings.HasPrefix(raw, prefix) {
		return raw[len(prefix):]
	}

	log.Println("❌ Invalid S3 URL:", raw)
	return ""
}
