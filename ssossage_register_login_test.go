package main

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hyperfyodor/ssosage_proto"
)

func TestRegisterLogin(t *testing.T) {

	ctx, suite := NewSuite(t)

	name := gofakeit.Name()
	password := gofakeit.Password(true, true, true, true, false, 20)

	_, err := suite.ssosageClient.Register(ctx, &ssosage_proto.RegisterRequest{Name: name, Password: password})

	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	resp2, err := suite.ssosageClient.Login(ctx, &ssosage_proto.LoginRequest{Name: name, Password: password})

	if err != nil {
		t.Fatalf("failed to login: %v", err)
	}

	if len(resp2.Token) == 0 {
		t.Fatalf("token %v is invlid", resp2.Token)
	}

	tokenParsed, err := jwt.Parse(resp2.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(APP_SECRET), nil
	})

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)

	if !ok {
		t.Fatal("failed to parse token")
	}

	if claims["name"] != name {
		t.Fatalf("%v != %v", claims["name"], name)
	}

}
