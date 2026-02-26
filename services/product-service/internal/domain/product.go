package domain

import (
	"github.com/google/uuid"
	common "github.com/pppestto/ecommerce-grpc/services/common/domain"
	commonerrors "github.com/pppestto/ecommerce-grpc/services/common/errors"
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       common.Money
	Stock       int32
}

func NewProduct(name, description string, price common.Money, stock int32) (*Product, error) {
	if name == "" {
		return nil, commonerrors.ErrInvalidArgument
	}
	if price.Amount <= 0 {
		return nil, commonerrors.ErrInvalidArgument
	}
	if stock < 0 {
		return nil, commonerrors.ErrInvalidArgument
	}

	return &Product{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Price:       price,
		Stock:       stock,
	}, nil
}
