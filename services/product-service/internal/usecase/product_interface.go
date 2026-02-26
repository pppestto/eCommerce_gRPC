package usecase

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"
)

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	List(ctx context.Context, page, size int, category string) ([]*domain.Product, int, error)
	Save(ctx context.Context, product *domain.Product) error
}
