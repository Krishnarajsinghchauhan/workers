package main

import (
	"log"
	"esign-worker/internal"
)

func main() {
	log.Println("eSign Worker Started")

	internal.InitSQS()
	internal.InitS3()
	internal.InitRedis()

	internal.ListenToQueue()
}
