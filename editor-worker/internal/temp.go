package internal

import (
	"os"
	"path/filepath"
	"time"
)

func TempName(prefix, ext string) string {
	return filepath.Join("/tmp", prefix+"_"+time.Now().Format("150405.000000")+ext)
}

func DeleteFile(path string) {
	os.Remove(path)
}
