package interfaces

import (
	"context"
	"ssosage/internal/models"
)

type ClientSaver interface {
	SaveClient(ctx context.Context, name string, passwordHash []byte) (int64, error)
}

type ClientProvider interface {
	Client(ctx context.Context, name string) (models.Client, error)
}

type AppSaver interface {
	SaveApp(ctx context.Context, name string, secret string, roles string) (int64, error)
}

type AppProvider interface {
	App(ctx context.Context, name string) (models.App, error)
}

type PasswordHasher interface {
	Hash(password string) ([]byte, error)
	Compare(hash []byte, password string) (bool, error)
}
