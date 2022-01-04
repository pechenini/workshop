package kafka

import (
	"context"
	"encoding/json"
	"github.com/segmentio/kafka-go"
	"strconv"
	"workshop/internal/todo"
)

type Publisher struct {
	writer *kafka.Writer
}

func NewPublisher(writer *kafka.Writer) *Publisher {
	return &Publisher{writer: writer}
}

func (p *Publisher) Publish(ctx context.Context, event todo.Event) error {
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.FormatInt(event.Todo.Id, 10)),
		Value: eventBytes,
	})
}
