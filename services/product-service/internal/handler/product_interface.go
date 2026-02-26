package handler

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"
)

type ProductService interface {
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
	ListProducts(ctx context.Context, page, size int, category string) ([]*domain.Product, int, error)
	CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error)
}
