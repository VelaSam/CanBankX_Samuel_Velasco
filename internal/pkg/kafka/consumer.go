package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
	"gorm.io/gorm"
)
func StartConsumer(db *gorm.DB, broker string, topic string, groupID string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	log.Printf("Starting Kafka consumer for topic %s on broker %s...", topic, broker)

	go func() {
		defer r.Close()
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				log.Printf("Error reading message from Kafka: %v", err)
				break
			}
			
			//on essye d run ca a travers notre bbc logic 
			ProcessBBCEvent(db, m.Value)
		}
	}()
}
