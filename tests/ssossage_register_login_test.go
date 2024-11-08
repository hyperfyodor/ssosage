package tests

import (
	"ssosage/tests/suite"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt"
	"github.com/hyperfyodor/ssosage_proto"
)

const APP_SECRET = "some_secret"

func TestRegisterLogin(t *testing.T) {
	ctx, suite := suite.NewSuite(t)

	appName := gofakeit.AppName()

	_, err := suite.SsosageClient.RegisterApp(
		ctx,
		&ssosage_proto.RegisterAppRequest{
			AppName:   appName,
			AppSecret: APP_SECRET,
			Roles:     []string{"user", "admin"},
		},
	)

	if err != nil {
		t.Fatalf("failed to register an app: %v", err)
	}

	t.Logf("created an app : %v", appName)

	clientName := gofakeit.AppName()
	password := gofakeit.Password(true, true, true, true, false, 20)

	_, err = suite.SsosageClient.RegisterClient(
		ctx,
		&ssosage_proto.RegisterClientRequest{
			ClientName: clientName,
			Password:   password,
		},
	)

	if err != nil {
		t.Fatalf("failed to register a client: %v", err)
	}

	t.Logf("created a client %v", clientName)

	role := "user"

	resp, err := suite.SsosageClient.GenerateToken(
		ctx,
		&ssosage_proto.GenerateTokenRequest{
			ClientName: clientName,
			Password:   password,
			AppName:    appName,
			Role:       role,
		},
	)

	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	t.Logf("generated a token for client %v, app %v, role %v", clientName, appName, role)

	tokenParsed, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(APP_SECRET), nil
	})

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)

	if !ok {
		t.Fatal("failed to parse token")
	}

	if claims["role"] != role || claims["client_name"] != clientName || claims["app_name"] != appName || !tokenParsed.Valid {
		t.Fatalf("invalid token")
	}

}
