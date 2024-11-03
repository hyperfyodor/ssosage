package interfaces

import (
	"context"
	"ssosage/internal/models"
)

type UserSaver interface {
	SaveUser(ctx context.Context, name string, passwordHash []byte) error
}

type UserProvider interface {
	User(ctx context.Context, name string) (models.User, error)
}

type PasswordHasher interface {
	Hash(password string) ([]byte, error)
	Compare(hash []byte, password string) (bool, error)
}
