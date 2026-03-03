package kafka

import (
	"context"
)

type outboxStoreFuncs struct {
	get  func(ctx context.Context, limit int) ([]OutboxRow, error)
	mark func(ctx context.Context, id string) error
}

func NewOutboxStore(
	get func(ctx context.Context, limit int) ([]OutboxRow, error),
	mark func(ctx context.Context, id string) error,
) OutboxStore {
	return &outboxStoreFuncs{get: get, mark: mark}
}

func (o *outboxStoreFuncs) GetUnpublishedOutbox(ctx context.Context, limit int) ([]OutboxRow, error) {
	return o.get(ctx, limit)
}

func (o *outboxStoreFuncs) MarkOutboxPublished(ctx context.Context, id string) error {
	return o.mark(ctx, id)
}
