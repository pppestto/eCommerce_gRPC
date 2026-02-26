package handler

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

// UserService - интерфейс бизнес логики
type UserService interface {
	CreateUser(ctx context.Context, email string) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
}
