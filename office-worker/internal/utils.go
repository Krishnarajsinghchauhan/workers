package internal

import (
	"os"
	"path/filepath"
	"time"
)

// TEMP FILE NAME
func TempName(prefix, ext string) string {
	return filepath.Join("/tmp", prefix+"_"+time.Now().Format("150405")+ext)
}

// DELETE FILE
func DeleteFile(path string) {
	os.Remove(path)
}
