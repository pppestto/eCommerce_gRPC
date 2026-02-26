package usecase

import (
	"context"
	"fmt"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

// UserUseCase - сервис бизнес логики
type UserService struct {
	repo     UserRepository
	eventBus UserEventBus
}

// NewUserService создаёт новый сервис пользователей
func NewUserService(repo UserRepository, eventBus UserEventBus) *UserService {
	return &UserService{
		repo:     repo,
		eventBus: eventBus,
	}
}

// CreateUser создаёт нового пользователя
func (s *UserService) CreateUser(ctx context.Context, email string) (*domain.User, error) {
	// Валидация и создание (domain logic)
	user, err := domain.NewUser(email)
	if err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	// Сохраняем в хранилище
	if err := s.repo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// Публикуем событие (асинхронно, не критично если упадёт)
	if err := s.eventBus.PublishUserCreated(ctx, user); err != nil {
		// Логируем, но не падаем
		fmt.Printf("failed to publish user created event: %v\n", err)
	}

	return user, nil
}

// GetUser получает пользователя по ID
func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	// Проверяем, что пользователь существует
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Удаляем из хранилища
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Публикуем событие
	if err := s.eventBus.PublishUserDeleted(ctx, id); err != nil {
		// Логируем, но не падаем
		fmt.Printf("failed to publish user deleted event: %v\n", err)
	}

	return nil
}
