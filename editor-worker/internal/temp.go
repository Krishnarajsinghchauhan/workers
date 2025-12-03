package internal

import (
	"os"
	"path/filepath"
	"time"
	"math/rand"
)

func TempName(prefix, ext string) string {
	return filepath.Join("/tmp", prefix+"_"+time.Now().Format("150405.000000")+ext)
}

func DeleteFile(path string) {
	os.Remove(path)
}

func RandString() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
