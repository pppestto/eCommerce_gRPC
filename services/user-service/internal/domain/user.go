package domain

import (
	"errors"
	"github.com/google/uuid"
	"strings"
)

type User struct {
	ID           string
	Email        string
	PasswordHash string
}

func NewUser(email string) (*User, error) {
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: "",
	}, nil
}

func validateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}

	if len(email) > 100 {
		return errors.New("email is too long (max 100 characters)")
	}

	if !strings.Contains(email, "@") {
		return errors.New("email must contain @")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return errors.New("invalid email format")
	}

	if parts[0] == "" {
		return errors.New("email local part cannot be empty")
	}

	if parts[1] == "" {
		return errors.New("email domain cannot be empty")
	}

	return nil
}
