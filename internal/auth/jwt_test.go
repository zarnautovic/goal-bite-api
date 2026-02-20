package auth

import (
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTManagerWithKeys(t *testing.T) {
	m, err := NewJWTManagerWithKeys("v2", map[string]string{
		"v1": "old-secret",
		"v2": "new-secret",
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	token, err := m.Generate(42)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("new-secret"), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("parse raw token: %v", err)
	}
	if kid, _ := parsed.Header["kid"].(string); kid != "v2" {
		t.Fatalf("expected kid v2, got %v", parsed.Header["kid"])
	}

	userID, err := m.Parse(token)
	if err != nil {
		t.Fatalf("parse managed token: %v", err)
	}
	if userID != 42 {
		t.Fatalf("expected user id 42, got %d", userID)
	}
}

func TestJWTManagerRejectsMissingOrUnknownKID(t *testing.T) {
	m, err := NewJWTManagerWithKeys("v1", map[string]string{
		"v1": "secret1",
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	claims := jwt.MapClaims{"sub": "1"}
	noKid := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	noKidToken, err := noKid.SignedString([]byte("secret1"))
	if err != nil {
		t.Fatalf("sign no-kid token: %v", err)
	}
	if _, err := m.Parse(noKidToken); err == nil {
		t.Fatalf("expected missing kid token to be rejected")
	}

	unknownKid := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	unknownKid.Header["kid"] = "v9"
	unknownKidToken, err := unknownKid.SignedString([]byte("secret1"))
	if err != nil {
		t.Fatalf("sign unknown-kid token: %v", err)
	}
	if _, err := m.Parse(unknownKidToken); err == nil {
		t.Fatalf("expected unknown kid token to be rejected")
	}
}

func TestNewJWTManagerWithKeysValidation(t *testing.T) {
	_, err := NewJWTManagerWithKeys("", map[string]string{"v1": "x"})
	if err == nil {
		t.Fatalf("expected error for empty active kid")
	}

	_, err = NewJWTManagerWithKeys("v1", nil)
	if err == nil {
		t.Fatalf("expected error for empty keyset")
	}

	_, err = NewJWTManagerWithKeys("v2", map[string]string{"v1": "x"})
	if err == nil || !strings.Contains(err.Error(), "active kid") {
		t.Fatalf("expected active kid error, got %v", err)
	}
}
