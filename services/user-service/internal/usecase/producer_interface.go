package usecase

import (
	"context"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

type UserEventBus interface {
	PublishUserCreated(ctx context.Context, user *domain.User) error
	PublishUserDeleted(ctx context.Context, userID string) error
}
