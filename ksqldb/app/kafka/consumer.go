package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
	topic    string
}

func NewConsumer(broker, topic, groupID string) (*Consumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	if err := c.Subscribe(topic, nil); err != nil {
		c.Close()
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	return &Consumer{
		consumer: c,
		topic:    topic,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context, handler func(Order)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				continue
			}

			var order Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				fmt.Printf("Failed to unmarshal: %v\n", err)
				continue
			}

			handler(order)
		}
	}
}

func (c *Consumer) Close() {
	c.consumer.Close()
}
