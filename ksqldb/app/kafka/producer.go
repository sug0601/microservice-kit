package kafka

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Order struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Product    string  `json:"product"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

type Producer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(broker, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Handle delivery reports in background
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Delivery failed: %v\n", ev.TopicPartition.Error)
				}
			}
		}
	}()

	return &Producer{
		producer: p,
		topic:    topic,
	}, nil
}

func (p *Producer) SendOrder(order Order) error {
	value, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	err = p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Key:            []byte(order.OrderID),
		Value:          value,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}

func (p *Producer) Flush() {
	p.producer.Flush(5000)
}

func (p *Producer) Close() {
	p.producer.Close()
}
