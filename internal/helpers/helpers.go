package helpers

import (
	"fmt"
	"log/slog"
	"ssosage/internal/models"
	"time"

	"github.com/golang-jwt/jwt"
)

func WrapErr(s string, e error) error {
	return fmt.Errorf("%s : %w", s, e)
}

func SlErr(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func NewToken(user models.User, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["name"] = user.Name
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte("some_secret")) // TODO - secret per app

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
