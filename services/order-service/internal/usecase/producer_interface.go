package usecase

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

type OrderEventBus interface {
	PublishOrderCreated(ctx context.Context, order *domain.Order) error
	PublishOrderStatusChanged(ctx context.Context, order *domain.Order) error
}
