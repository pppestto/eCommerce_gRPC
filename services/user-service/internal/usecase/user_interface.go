package usecase

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashed, plain string) bool
}

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}
