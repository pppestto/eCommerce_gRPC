package kafka

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

type KafkaEventBus struct {
	writer *kafka.Writer
}

// New создаёт новый Kafka event bus
func New() (*KafkaEventBus, error) {
	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
	topic := getEnv("KAFKA_USER_TOPIC", "user-events")
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
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
		return errors.Wrap(err, "failed to marshal event")
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
		return errors.Wrap(err, "failed to write message")
	}

	return nil
}

func (k *KafkaEventBus) PublishUserDeleted(ctx context.Context, userID string) error {
	event := UserDeletedEvent{
		UserID: userID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
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
		return errors.Wrap(err, "failed to write message")
	}

	return nil
}

// Close закрывает соединение с Kafka
func (k *KafkaEventBus) Close() error {
	return k.writer.Close()
}
