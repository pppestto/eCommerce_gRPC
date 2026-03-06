package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type OutboxRow struct {
	ID        string
	EventType string
	Payload   []byte
}

type OutboxStore interface {
	GetUnpublishedOutbox(ctx context.Context, limit int) ([]OutboxRow, error)
	MarkOutboxPublished(ctx context.Context, id string) error
}

func (k *OrderKafkaProducer) PublishMessage(ctx context.Context, key []byte, eventType string, payload []byte) error {
	msg := kafka.Message{
		Key:   key,
		Value: payload,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(eventType)},
		},
	}
	return k.writer.WriteMessages(ctx, msg)
}

func OutboxRelay(ctx context.Context, store OutboxStore, producer *OrderKafkaProducer, batchSize int, interval time.Duration, logger *slog.Logger) {
	if batchSize <= 0 {
		batchSize = 50
	}
	if interval <= 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := relayBatch(ctx, store, producer, batchSize, logger); err != nil {
				logger.Error("outbox relay: batch failed", "error", err)
			}
		}
	}
}

func relayBatch(ctx context.Context, store OutboxStore, producer *OrderKafkaProducer, limit int, logger *slog.Logger) error {
	rows, err := store.GetUnpublishedOutbox(ctx, limit)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if err := producer.PublishMessage(ctx, []byte(row.ID), row.EventType, row.Payload); err != nil {
			logger.Error("outbox relay: failed to publish", "error", err, "outbox_id", row.ID)
			continue
		}

		if err := store.MarkOutboxPublished(ctx, row.ID); err != nil {
			logger.Error("outbox relay: failed to mark published", "error", err, "outbox_id", row.ID)
		}
	}

	return nil
}
