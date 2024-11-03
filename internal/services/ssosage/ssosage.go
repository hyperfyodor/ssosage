package ssosage

import (
	"context"
	"errors"
	"log/slog"
	"ssosage/internal/helpers"
	"ssosage/internal/interfaces"
	"ssosage/internal/storage"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Ssosage struct {
	log          *slog.Logger
	userProvider interfaces.UserProvider
	userSaver    interfaces.UserSaver
	hasher       interfaces.PasswordHasher
}

func New(
	log *slog.Logger,
	userProvider interfaces.UserProvider,
	userSaver interfaces.UserSaver,
	hasher interfaces.PasswordHasher,
) *Ssosage {

	if log == nil {
		panic("can't create Ssosage structure - *logger is nil ")
	}

	return &Ssosage{
		log:          log,
		userProvider: userProvider,
		userSaver:    userSaver,
		hasher:       hasher,
	}

}

func (s *Ssosage) RegisterNewUser(ctx context.Context, name string, password string) error {

	const op = "srvices.ssosage.RegisterNewUser"

	log := s.logWith(op, name)
	log.Info("registering user")

	passwordHash, err := s.hasher.Hash(password)

	if err != nil {
		log.Error("failed to generate hash", helpers.SlErr(err))

		return helpers.WrapErr(op, err)
	}

	err = s.userSaver.SaveUser(ctx, name, passwordHash)

	if err != nil {

		log.Error("failed to save user", helpers.SlErr(err))

		return helpers.WrapErr(op, err)
	}

	return nil

}

func (s *Ssosage) Login(ctx context.Context, name string, password string) (string, error) {

	const op = "services.ssosage.Login"

	log := s.logWith(op, name)

	log.Info("logging in")

	user, err := s.userProvider.User(ctx, name)

	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", helpers.SlErr(err))

			return "", ErrInvalidCredentials
		}

		log.Error("failed to get user", helpers.SlErr(err))

		return "", helpers.WrapErr(op, err)
	}

	ok, err := s.hasher.Compare(user.PasswordHash, password)

	if err != nil {
		log.Error("failed to campare hash", helpers.SlErr(err))
	}

	if !ok {
		log.Info("invalid credentials", helpers.SlErr(err))

		return "", ErrInvalidCredentials
	}

	token, err := helpers.NewToken(user, 5*time.Hour) // TODO - duration should be set on the app side, per role?

	if err != nil {
		log.Info("failed to generate token", helpers.SlErr(err))

		return "", helpers.WrapErr(op, err)
	}

	return token, nil

}

func (s *Ssosage) logWith(op string, name string) *slog.Logger {
	return s.log.With(
		slog.String("op", op),
		slog.String("name", name),
	)
}
