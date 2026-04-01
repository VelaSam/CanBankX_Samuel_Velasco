package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"canbankx/internal/model"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Consumer struct {
	consumer *kafka.Consumer
	topic    string
}

func NewConsumer() (*Consumer, error) {
	host := os.Getenv("KAFKA_HOST")
	if host == "" {
		host = "localhost:9092" // defaults
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "transactions"
	}

	groupID := os.Getenv("KAFKA_GROUP_ID")
	if groupID == "" {
		groupID = "canbankx-group"
	}

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": host,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	return &Consumer{consumer: c, topic: topic}, nil
}

func (c *Consumer) Consume(ctx context.Context, handle func(model.Transaction) error) error {
	slog.Info("Starting Kafka consumer...", "topic", c.topic)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				var t model.Transaction
				if err := json.Unmarshal(e.Value, &t); err != nil {
					slog.Error("Failed to unmarshal transaction", "error", err)
					continue
				}

				slog.Debug("Kafka message received",
					"topic", *e.TopicPartition.Topic,
					"partition", int(e.TopicPartition.Partition),
					"offset", int64(e.TopicPartition.Offset))

				if err := handle(t); err != nil {
					slog.Error("Error handling message", "error", err)
				}
			case kafka.Error:
				if e.IsFatal() {
					return e
				}
				slog.Warn("Kafka consumer error", "error", e)
			}
		}
	}
}

func (c *Consumer) Close() {
	if err := c.consumer.Close(); err != nil {
		slog.Error("Failed to close kafka consumer", "error", err)
	}
}
