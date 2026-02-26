package kafka

import (
	"context"
	"encoding/json"
	"time"

	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type OrderKafkaProducer struct {
	writer *kafka.Writer
}

func NewOrderKafkaProducer(address []string, topic string) (*OrderKafkaProducer, error) {
	if len(address) == 0 {
		return nil, commonerrors.ErrInvalidArgument
	}
	if topic == "" {
		return nil, commonerrors.ErrInvalidArgument
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(address...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	return &OrderKafkaProducer{writer: writer}, nil
}

func (k *OrderKafkaProducer) PublishOrderCreated(ctx context.Context, order *domain.Order) error {
	jsonOrder, err := json.Marshal(order)
	if err != nil {
		return errors.Wrap(err, "failed to marshal order")
	}

	message := kafka.Message{
		Key:   []byte(order.ID.String()),
		Value: jsonOrder,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte("order.created")},
		},
	}

	if err := k.writer.WriteMessages(ctx, message); err != nil {
		return errors.Wrap(err, "failed to send order to kafka")
	}

	return nil
}

func (k *OrderKafkaProducer) PublishOrderStatusChanged(ctx context.Context, order *domain.Order) error {
	jsonOrder, err := json.Marshal(order)
	if err != nil {
		return errors.Wrap(err, "failed to marshal order")
	}

	message := kafka.Message{
		Key:   []byte(order.ID.String()),
		Value: jsonOrder,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte("order.status_changed")},
		},
	}

	if err := k.writer.WriteMessages(ctx, message); err != nil {
		return errors.Wrap(err, "failed to send order to kafka")
	}

	return nil
}
