package usecase

import (
	"context"
	"encoding/json"

	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

type OutboxEventSpec interface {
	AggregateType() string
	AggregateID() string
	EventType() string
	Payload() json.RawMessage
}

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	SaveOrderWithOutbox(ctx context.Context, order *domain.Order, event OutboxEventSpec) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error)
	UpdateStatusWithOutbox(ctx context.Context, id string, status domain.OrderStatus, event OutboxEventSpec) (*domain.Order, error)
}
