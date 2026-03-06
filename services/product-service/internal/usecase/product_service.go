package usecase

import (
	"context"

	"github.com/pkg/errors"
	"github.com/pppestto/ecommerce-grpc/services/product-service/internal/domain"
)

type ProductService struct {
	eventBus ProductEventBus
	repo     ProductRepository
}

func NewProductService(repo ProductRepository, eventBus ProductEventBus) *ProductService {
	return &ProductService{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get product by id")
	}

	return product, nil
}

func (s *ProductService) ListProducts(ctx context.Context, page, size int, category string) ([]*domain.Product, int, error) {
	products, total, err := s.repo.List(ctx, page, size, category)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to list products")
	}
	return products, total, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, product domain.Product) (*domain.Product, error) {
	createdProduct, err := domain.NewProduct(product.Name, product.Description, product.Price, product.Stock)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new product")
	}

	if err = s.repo.Save(ctx, createdProduct); err != nil {
		return nil, errors.Wrap(err, "failed to save product")
	}

	if err = s.eventBus.PublishProductCreated(ctx, createdProduct); err != nil {
		return nil, errors.Wrap(err, "failed to publish product created")
	}

	return createdProduct, nil
}
