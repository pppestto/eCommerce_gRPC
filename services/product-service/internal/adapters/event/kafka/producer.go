package kafka

import (
	"context"
	"encoding/json"
	"time"

	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

type ProductKafkaProducer struct {
	writer *kafka.Writer
}

func NewProductKafkaProducer(address []string, topic string) (*ProductKafkaProducer, error) {
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

	return &ProductKafkaProducer{writer: writer}, nil
}

func (k *ProductKafkaProducer) PublishProductCreated(ctx context.Context, product *domain.Product) error {
	jsonProduct, err := json.Marshal(product)
	if err != nil {
		return errors.Wrap(err, "failed to marshal product")
	}

	message := kafka.Message{
		Key:   []byte(product.ID.String()),
		Value: jsonProduct,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte("product.created")},
		},
	}

	if err := k.writer.WriteMessages(ctx, message); err != nil {
		return errors.Wrap(err, "failed to send product to kafka")
	}

	return nil
}
