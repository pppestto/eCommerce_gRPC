package usecase

import (
	"context"
	"fmt"

	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/domain"
)

type UserService struct {
	repo           UserRepository
	eventBus       UserEventBus
	passwordHasher PasswordHasher
}

func NewUserService(repo UserRepository, eventBus UserEventBus, passwordHasher PasswordHasher) *UserService {
	return &UserService{
		repo:           repo,
		eventBus:       eventBus,
		passwordHasher: passwordHasher,
	}
}

func (s *UserService) CreateUser(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := domain.NewUser(email)
	if err != nil {
		return nil, fmt.Errorf("invalid user data: %w", err)
	}

	if password != "" {
		hash, err := s.passwordHasher.Hash(password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = hash
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	if err := s.eventBus.PublishUserCreated(ctx, user); err != nil {
		fmt.Printf("failed to publish user created event: %v\n", err)
	}

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if user.PasswordHash == "" {
		return nil, fmt.Errorf("invalid email or password")
	}

	if !s.passwordHasher.Compare(user.PasswordHash, password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := s.eventBus.PublishUserDeleted(ctx, id); err != nil {
		fmt.Printf("failed to publish user deleted event: %v\n", err)
	}

	return nil
}
