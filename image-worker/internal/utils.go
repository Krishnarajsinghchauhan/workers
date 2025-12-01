package internal

import (
	"os"
	"path/filepath"
	"time"
)

// Generate temp filename inside /tmp
func TempFile(prefix, ext string) string {
	name := prefix + "_" + time.Now().Format("150405") + ext
	return filepath.Join("/tmp", name)
}

// Safe delete
func DeleteFile(path string) {
	_ = os.Remove(path)
}
