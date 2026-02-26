package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

type KafkaEventBus struct {
	writer *kafka.Writer
}

// New создаёт новый Kafka event bus
func New() (*KafkaEventBus, error) {
	// TODO: Заменить на реальный адрес Kafka из конфига
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "user-events",
	})

	return &KafkaEventBus{
		writer: writer,
	}, nil
}

type UserDeletedEvent struct {
	UserID string `json:"user_id"`
}

func (k *KafkaEventBus) PublishUserCreated(ctx context.Context, user *domain.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(user.ID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte("user.created")},
		},
	}

	err = k.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

func (k *KafkaEventBus) PublishUserDeleted(ctx context.Context, userID string) error {
	event := UserDeletedEvent{
		UserID: userID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	message := kafka.Message{
		Key:   []byte(userID),
		Value: data,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte("user.deleted")},
		},
	}

	err = k.writer.WriteMessages(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Close закрывает соединение с Kafka
func (k *KafkaEventBus) Close() error {
	return k.writer.Close()
}
