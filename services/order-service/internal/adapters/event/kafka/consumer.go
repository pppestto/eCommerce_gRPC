package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/segmentio/kafka-go"
)

// OrderEvent — DTO для события заказа из Kafka (ID как string для JSON)
type OrderEvent struct {
	ID     string         `json:"ID"`
	UserID string         `json:"UserID"`
	Items  []OrderItemEv  `json:"Items"`
	Total  MoneyEv        `json:"Total"`
	Status int            `json:"Status"`
}

type OrderItemEv struct {
	ProductID string   `json:"ProductID"`
	Quantity  int      `json:"Quantity"`
	Price     MoneyEv  `json:"Price"`
}

type MoneyEv struct {
	Amount   int64  `json:"Amount"`
	Currency string `json:"Currency"`
}

// OrderEventHandler вызывается для каждого полученного события заказа
type OrderEventHandler func(ctx context.Context, event OrderEvent, eventType string) error

type OrderKafkaConsumer struct {
	reader *kafka.Reader
}

func NewOrderKafkaConsumer(brokers []string, topic, groupID string) (*OrderKafkaConsumer, error) {
	if len(brokers) == 0 {
		return nil, commonerrors.ErrInvalidArgument
	}
	if topic == "" {
		return nil, commonerrors.ErrInvalidArgument
	}
	if groupID == "" {
		return nil, commonerrors.ErrInvalidArgument
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		Topic:          topic,
		GroupID:        groupID,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &OrderKafkaConsumer{reader: reader}, nil
}

// Consume читает сообщения и вызывает handler для каждого. Блокирует до отмены ctx.
func (c *OrderKafkaConsumer) Consume(ctx context.Context, handler OrderEventHandler) {
	defer c.reader.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("order consumer: read message: %v", err)
				return
			}

			eventType := getHeader(msg.Headers, "event-type")
			var ev OrderEvent
			if err := json.Unmarshal(msg.Value, &ev); err != nil {
				log.Printf("order consumer: unmarshal: %v", err)
				continue
			}

			if handler != nil {
				if err := handler(ctx, ev, eventType); err != nil {
					log.Printf("order consumer: handler: %v", err)
				}
			}
		}
	}
}

func getHeader(headers []kafka.Header, key string) string {
	for _, h := range headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}
