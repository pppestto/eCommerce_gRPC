package usecase

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"
)

type ProductEventBus interface {
	PublishProductCreated(ctx context.Context, product *domain.Product) error
}
