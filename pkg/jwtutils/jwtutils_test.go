package jwtutils_test

import (
	"testing"

	"github.com/B0TMirage/avito-assignment-winter-2025.git/pkg/jwtutils"
)

func TestCreateToken(t *testing.T) {
	username := "testuser"
	secretKey := "mysecretkey"

	tokenString, err := jwtutils.CreateToken(username, secretKey)
	if err != nil {
		t.Fatalf("got %v, want no error", err)
	}

	if tokenString == "" {
		t.Fatal("want non-empty token string")
	}
}

func TestValidateJWT(t *testing.T) {
	username := "testuser"
	secretKey := "mysecretkey"

	tokenString, err := jwtutils.CreateToken(username, secretKey)
	if err != nil {
		t.Fatalf("failed to create token")
	}

	claims, err := jwtutils.ValidateJWT(tokenString, secretKey)
	if err != nil {
		t.Fatalf("got %v, want no error", err)
	}

	if claims["username"] != username {
		t.Fatalf("got %v, expected username %v", claims["username"], username)
	}

	if claims["ttl"] == nil {
		t.Fatal("want ttl claim to be present")
	}
}

func TestValidateJWTInvalidToken(t *testing.T) {
	invalidToken := "invalid.token.string"
	secretKey := "mysecretkey"

	_, err := jwtutils.ValidateJWT(invalidToken, secretKey)
	if err == nil {
		t.Fatal("got none, want error for invalid token")
	}
}

func TestValidateJWTInvalidSigningMethod(t *testing.T) {
	username := "testuser"
	secretKey := "mysecretkey"

	tokenString, _ := jwtutils.CreateToken(username, secretKey)

	tamperedToken := tokenString + "tampered"

	_, err := jwtutils.ValidateJWT(tamperedToken, secretKey)
	if err == nil {
		t.Fatal("got none, want error for tampered token")
	}
}
