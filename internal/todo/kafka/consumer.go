package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Consumer struct {
	reader *kafka.Reader
	logger logrus.FieldLogger
}

func NewConsumer(reader *kafka.Reader, logger logrus.FieldLogger) *Consumer {
	return &Consumer{reader: reader, logger: logger}
}

func (c Consumer) Consume() error {
	for {
		msg, err := c.reader.ReadMessage(context.Background())
		if err != nil {
			return err
		}
		c.logger.WithFields(map[string]interface{}{
			"topic": msg.Topic,
			"key": string(msg.Key),
			"value": string(msg.Value),
			"partition": msg.Partition,
			"time": msg.Time.String(),
		}).Info("Message received")
	}
}