package usecase

import (
	"context"
	"log/slog"

	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "invalid user data")
	}

	if password != "" {
		hash, err := s.passwordHasher.Hash(password)
		if err != nil {
			return nil, errors.Wrap(err, "failed to hash password")
		}
		user.PasswordHash = hash
	}

	if err := s.repo.Save(ctx, user); err != nil {
		return nil, errors.Wrap(err, "failed to save user")
	}

	if err := s.eventBus.PublishUserCreated(ctx, user); err != nil {
		slog.Warn("failed to publish user created event", "error", err, "user_id", user.ID)
	}

	return user, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "user not found")
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*domain.User, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "invalid email or password")
	}

	if user.PasswordHash == "" {
		return nil, errors.New("invalid email or password")
	}

	if !s.passwordHasher.Compare(user.PasswordHash, password) {
		return nil, errors.New("invalid email or password")
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return errors.Wrap(err, "user not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return errors.Wrap(err, "failed to delete user")
	}

	if err := s.eventBus.PublishUserDeleted(ctx, id); err != nil {
		slog.Warn("failed to publish user deleted event", "error", err, "user_id", id)
	}

	return nil
}
