package auth

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/pkg/errors"
	"github.com/pppestto/ecommerce-grpc/services/user-service/internal/usecase"
)

type BcryptHasher struct{}

var _ usecase.PasswordHasher = (*BcryptHasher)(nil)

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (b *BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.Wrap(err, "failed to hash password")
	}
	return string(hash), nil
}

func (b *BcryptHasher) Compare(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
