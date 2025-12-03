package internal

import (
	"path/filepath"
	"time"
)

// EXACT FUNCTION NAME NEEDED BY OCR
func TempName(prefix, ext string) string {
	return filepath.Join("/tmp", prefix+"_"+time.Now().Format("150405")+ext)
}

func TempFile(prefix, ext string) string {
	return TempName(prefix, ext)
}

func DeleteFile(path string) {
	_ = os.Remove(path)
}
