package usecase

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

// UserRepository - интерфейс для работы с хранилищем
type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}
