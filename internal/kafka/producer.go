package kafka

import (
	"context"
	"github.com/segmentio/kafka-go"
	"log"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &Producer{writer: writer}
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (p *Producer) Publish(ctx context.Context, key []byte, value []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: value,
	}
	err := p.writer.WriteMessages(ctx, msg)
	if err != nil {
		log.Printf("failed to publish message to kafka: %v", err)
	}
	return err
}
