package usecase

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

type OrderRepository interface {
	Save(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id string) (*domain.Order, error)
	UpdateStatus(ctx context.Context, id string, status domain.OrderStatus) (*domain.Order, error)
}
