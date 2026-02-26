package domain

import (
	"errors"

	"github.com/google/uuid"
	common "github.com/pppestto/ecommerce-grpc/services/common/domain"
)

var (
	ErrInvalidQuantity = errors.New("quantity must be positive")
	ErrInvalidPrice    = errors.New("price must be valid")
)

type OrderItem struct {
	ProductID uuid.UUID
	Quantity  int
	Price     common.Money
}

// NewOrderItem создаёт позицию заказа с валидацией
func NewOrderItem(productID uuid.UUID, quantity int, price common.Money) (*OrderItem, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if price.Currency == "" || price.Amount < 0 {
		return nil, ErrInvalidPrice
	}

	return &OrderItem{
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}, nil
}
