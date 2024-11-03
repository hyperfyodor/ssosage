package tests

import (
	"ssosage/tests/suite"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hyperfyodor/ssosage_proto"
)

const APP_SECRET = "some_secret"

func TestRegisterLogin(t *testing.T) {

	ctx, suite := suite.NewSuite(t)

	name := gofakeit.Name()
	password := gofakeit.Password(true, true, true, true, false, 20)

	_, err := suite.SsosageClient.Register(ctx, &ssosage_proto.RegisterRequest{Name: name, Password: password})

	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	resp2, err := suite.SsosageClient.Login(ctx, &ssosage_proto.LoginRequest{Name: name, Password: password})

	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	if len(resp2.Token) == 0 {
		t.Fatalf("token %v is invlid", resp2.Token)
	}

	tokenParsed, err := jwt.Parse(resp2.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(APP_SECRET), nil
	})

	if err != nil {
		t.Fatal("failed to parse token")
	}

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)

	if !ok {
		t.Fatal("failed to parse token")
	}

	if claims["name"] != name {
		t.Fatalf("%v != %v", claims["name"], name)
	}

}
