package ssosage

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"ssosage/internal/helpers"
	"ssosage/internal/interfaces"
	"ssosage/internal/models"
	"ssosage/internal/storage"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidApp         = errors.New("invalid app")
	ErrInvalidRole        = errors.New("invalid role")
)

type Ssosage struct {
	log            *slog.Logger
	clientSaver    interfaces.ClientSaver
	clientProvider interfaces.ClientProvider
	appSaver       interfaces.AppSaver
	appProvider    interfaces.AppProvider
	hasher         interfaces.PasswordHasher
}

func New(
	log *slog.Logger,
	clientSaver interfaces.ClientSaver,
	clientProvider interfaces.ClientProvider,
	appSaver interfaces.AppSaver,
	appProvider interfaces.AppProvider,
	hasher interfaces.PasswordHasher,
) *Ssosage {

	if log == nil {
		panic("can't create Ssosage structure - *logger is nil ")
	}

	return &Ssosage{
		log:            log,
		clientSaver:    clientSaver,
		clientProvider: clientProvider,
		appSaver:       appSaver,
		appProvider:    appProvider,
		hasher:         hasher,
	}

}

func (s *Ssosage) RegisterNewClient(ctx context.Context, name string, password string) (int64, error) {

	const op = "srvices.ssosage.RegisterNewClient"

	log := s.logWith(op, name)
	log.Info("registering client")

	passwordHash, err := s.hasher.Hash(password)

	if err != nil {
		log.Error("failed to generate hash", helpers.SlErr(err))

		return 0, helpers.WrapErr(op, err)
	}

	id, err := s.clientSaver.SaveClient(ctx, name, passwordHash)

	if err != nil {

		log.Error("failed to save user", helpers.SlErr(err))

		return 0, helpers.WrapErr(op, err)
	}

	return id, nil

}

func (s *Ssosage) RegisterNewApp(ctx context.Context, name string, secret string, roles string) (int64, error) {

	const op = "srvices.ssosage.RegisterNewApp"

	log := s.logWith(op, name)
	log.Info("registering app")

	id, err := s.appSaver.SaveApp(ctx, name, secret, roles)

	if err != nil {

		log.Error("failed to save app", helpers.SlErr(err))

		return 0, helpers.WrapErr(op, err)
	}

	return id, nil

}

func (s *Ssosage) GenerateToken(ctx context.Context, clientName string, password string, appName string, role string) (string, error) {

	const op = "services.ssosage.GenerateToken"

	log := s.logWith(op, clientName)

	log.Info("logging in")

	client, err := s.clientProvider.Client(ctx, clientName)

	if err != nil {
		if errors.Is(err, storage.ErrClientNotFound) {
			log.Warn("client not found", helpers.SlErr(err))

			return "", helpers.WrapErr(op, ErrInvalidCredentials)
		}

		log.Error("failed to get client", helpers.SlErr(err))

		return "", helpers.WrapErr(op, err)
	}

	ok, err := s.hasher.Compare(client.PasswordHash, password)

	if err != nil {
		log.Error("failed to campare hash", helpers.SlErr(err))

		return "", helpers.WrapErr(op, err)
	}

	if !ok {
		log.Info("invalid credentials", helpers.SlErr(err))

		return "", helpers.WrapErr(op, ErrInvalidCredentials)
	}

	app, err := s.appProvider.App(ctx, appName)

	if err != nil {
		if errors.Is(err, storage.ErrAppNotFound) {
			log.Warn("app not found", helpers.SlErr(err))

			return "", helpers.WrapErr(op, ErrInvalidApp)
		}

		return "", helpers.WrapErr(op, err)
	}

	token, err := s.newToken(client, app, role, 5*time.Hour)

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

func (s *Ssosage) newToken(client models.Client, app models.App, role string, duration time.Duration) (string, error) {

	const op = "services.ssosage.newToken"

	log := s.logWith(op, client.Name)

	token := jwt.New(jwt.SigningMethodHS256)

	app_roles := strings.Split(app.Roles, ",")

	if ok := slices.ContainsFunc[[]string, string](app_roles, func(r string) bool { return r == role }); !ok {
		log.Warn("client wants a role that is absent in that app", "role", role, "app", app.Name)

		return "", helpers.WrapErr(op, ErrInvalidRole)
	}

	claims := token.Claims.(jwt.MapClaims)

	claims["client_id"] = client.ID
	claims["client_name"] = client.Name
	claims["app_name"] = app.Name
	claims["role"] = role
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(app.Secret))

	if err != nil {

		log.Error("failed to create signed token string", helpers.SlErr(err))

		return "", helpers.WrapErr(op, err)
	}

	return tokenString, nil
}
