package main

import (
	"log"
	"editor-worker/internal"
)

func main() {
	log.Println("Editor Worker Started")

	internal.InitSQS()
	internal.InitS3()
	internal.InitRedis()

	internal.ListenToQueue()
}
