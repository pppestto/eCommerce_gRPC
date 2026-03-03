package handler

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

type UserService interface {
	CreateUser(ctx context.Context, email, password string) (*domain.User, error)
	GetUser(ctx context.Context, id string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (*domain.User, error)
	DeleteUser(ctx context.Context, id string) error
}
