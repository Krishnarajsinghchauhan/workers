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

