package main

import (
	"log"
	"combine-worker/internal"
)

func main() {
	log.Println("Combine Worker Started")

	internal.InitSQS()
	internal.InitS3()
	internal.InitRedis()

	internal.ListenToQueue()
}
