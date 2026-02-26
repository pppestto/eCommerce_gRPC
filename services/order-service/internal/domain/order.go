package domain

import (
	"errors"

	"github.com/google/uuid"
	common "github.com/pppestto/ecommerce-grpc/services/common/domain"
)

var (
	ErrEmptyItems   = errors.New("order must have at least one item")
	ErrNilTotal     = errors.New("order total cannot be nil")
	ErrInvalidTotal = errors.New("order total must match sum of items")
)

type Order struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Items  []OrderItem
	Total  common.Money
	Status OrderStatus
}

// NewOrder создаёт заказ с валидацией.
// total должен соответствовать сумме items (можно проверить в use case при необходимости).
func NewOrder(userID uuid.UUID, items []OrderItem, total common.Money, status OrderStatus) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyItems
	}
	// Проверяем что total != nil — common.Money не pointer, проверяем через Amount
	if total.Currency == "" {
		return nil, ErrNilTotal
	}

	return &Order{
		ID:     uuid.New(),
		UserID: userID,
		Items:  items,
		Total:  total,
		Status: status,
	}, nil
}
