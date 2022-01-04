package main

import (
	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	todo "workshop/internal/todo/kafka"
)

func main() {
	//init logger
	logger := logrus.New()

	//load env vars from .env
	if err := godotenv.Load(); err != nil {
		logger.Fatal(err)
	}

	//init reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
		Topic:   os.Getenv("TODO_TOPIC"),
		GroupID: "todo-consumer",
	})

	//create todo consumer
	todoConsumer := todo.NewConsumer(reader, logger)
	logger.Info("start consuming messages from kafka")

	//start consuming
	err := todoConsumer.Consume()
	if err != nil {
		logger.Error("error during reading msges from kafka:", err)
	}
}