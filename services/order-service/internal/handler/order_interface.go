package handler

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

// OrderService — интерфейс use case. Handler конвертирует pb↔domain и вызывает его.
// Use case не зависит от protobuf.
type OrderService interface {
	CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error)
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error)
}
