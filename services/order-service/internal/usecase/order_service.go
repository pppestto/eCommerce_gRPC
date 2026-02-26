package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	common "github.com/pppestto/ecommerce-grpc/services/common/domain"
	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

type OrderService struct {
	eventBus OrderEventBus
	repo     OrderRepository
}

func NewOrderService(repo OrderRepository, eventBus OrderEventBus) *OrderService {
	return &OrderService{repo: repo, eventBus: eventBus}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []domain.OrderItem) (*domain.Order, error) {
	if len(items) == 0 {
		return nil, commonerrors.ErrInvalidArgument
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.Wrap(commonerrors.ErrInvalidArgument, "invalid user_id")
	}

	var totalAmount int64
	currency := ""
	for _, item := range items {
		if currency == "" {
			currency = item.Price.Currency
		}
		totalAmount += item.Price.Amount * int64(item.Quantity)
	}
	total, err := common.NewMoney(totalAmount, currency)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new money")
	}

	order, err := domain.NewOrder(userUUID, items, *total, domain.OrderStatusPending)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new order")
	}

	if err := s.repo.Save(ctx, order); err != nil {
		return nil, errors.Wrap(err, "failed to save order")
	}

	if err := s.eventBus.PublishOrderCreated(ctx, order); err != nil {
		return nil, errors.Wrap(err, "failed to publish order created")
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get order by id") // repo возвращает commonerrors.ErrNotFound при отсутствии заказа
	}
	return order, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, orderID string, status domain.OrderStatus) (*domain.Order, error) {
	order, err := s.repo.UpdateStatus(ctx, orderID, status)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update order status")
	}

	if err = s.eventBus.PublishOrderStatusChanged(ctx, order); err != nil {
		return nil, errors.Wrap(err, "failed to bublish order status changed")
	}

	return order, nil
}
