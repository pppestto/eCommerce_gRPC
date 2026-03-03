package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/segmentio/kafka-go"
)

type ProcessedEventsStore interface {
	IsProcessed(ctx context.Context, idempotencyKey string) (bool, error)
	MarkProcessed(ctx context.Context, idempotencyKey string) error
}

type OrderEvent struct {
	ID     string        `json:"ID"`
	UserID string        `json:"UserID"`
	Items  []OrderItemEv `json:"Items"`
	Total  MoneyEv       `json:"Total"`
	Status int           `json:"Status"`
}

type OrderItemEv struct {
	ProductID string  `json:"ProductID"`
	Quantity  int     `json:"Quantity"`
	Price     MoneyEv `json:"Price"`
}

type MoneyEv struct {
	Amount   int64  `json:"Amount"`
	Currency string `json:"Currency"`
}

type OrderEventHandler func(ctx context.Context, event OrderEvent, eventType string) error

type OrderKafkaConsumer struct {
	reader         *kafka.Reader
	processedStore ProcessedEventsStore
}

func NewOrderKafkaConsumer(brokers []string, topic, groupID string, processedStore ProcessedEventsStore) (*OrderKafkaConsumer, error) {
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

	return &OrderKafkaConsumer{
		reader:         reader,
		processedStore: processedStore,
	}, nil
}

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

			idempotencyKey := fmt.Sprintf("%s:%d:%d", msg.Topic, msg.Partition, msg.Offset)

			if c.processedStore != nil {
				processed, err := c.processedStore.IsProcessed(ctx, idempotencyKey)
				if err != nil {
					log.Printf("order consumer: is processed check: %v", err)
					continue
				}
				if processed {
					continue
				}
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
				} else if c.processedStore != nil {
					if err := c.processedStore.MarkProcessed(ctx, idempotencyKey); err != nil {
						log.Printf("order consumer: mark processed: %v", err)
					}
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
