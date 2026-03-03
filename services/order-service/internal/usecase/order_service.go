package usecase

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	common "github.com/pppestto/ecommerce-grpc/services/common/domain"
	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
	"github.com/pppestto/ecommerce-grpc/services/order-service/internal/domain"
)

type outboxEvent struct {
	aggregateType string
	aggregateID   string
	eventType     string
	payload       json.RawMessage
}

func (e outboxEvent) AggregateType() string    { return e.aggregateType }
func (e outboxEvent) AggregateID() string      { return e.aggregateID }
func (e outboxEvent) EventType() string        { return e.eventType }
func (e outboxEvent) Payload() json.RawMessage { return e.payload }

type OrderService struct {
	repo OrderRepository
}

func NewOrderService(repo OrderRepository) *OrderService {
	return &OrderService{repo: repo}
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

	payload, err := json.Marshal(order)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal order for outbox")
	}

	event := outboxEvent{
		aggregateType: "order",
		aggregateID:   order.ID.String(),
		eventType:     "order.created",
		payload:       payload,
	}

	if err := s.repo.SaveOrderWithOutbox(ctx, order, event); err != nil {
		return nil, errors.Wrap(err, "failed to save order with outbox")
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
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	order.Status = status
	payload, err := json.Marshal(order)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal order for outbox")
	}

	event := outboxEvent{
		aggregateType: "order",
		aggregateID:   order.ID.String(),
		eventType:     "order.status_changed",
		payload:       payload,
	}

	return s.repo.UpdateStatusWithOutbox(ctx, orderID, status, event)
}
