package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type JWTManager struct {
	activeKID string
	keys      map[string][]byte
}

func NewJWTManager(secret string) *JWTManager {
	return &JWTManager{
		activeKID: "v1",
		keys: map[string][]byte{
			"v1": []byte(secret),
		},
	}
}

func NewJWTManagerWithKeys(activeKID string, keys map[string]string) (*JWTManager, error) {
	active := strings.TrimSpace(activeKID)
	if active == "" {
		return nil, errors.New("active kid is required")
	}
	if len(keys) == 0 {
		return nil, errors.New("at least one jwt key is required")
	}

	out := make(map[string][]byte, len(keys))
	for kid, secret := range keys {
		k := strings.TrimSpace(kid)
		s := strings.TrimSpace(secret)
		if k == "" || s == "" {
			return nil, errors.New("invalid jwt key set")
		}
		out[k] = []byte(s)
	}
	if _, ok := out[active]; !ok {
		return nil, errors.New("active kid not found in key set")
	}

	return &JWTManager{
		activeKID: active,
		keys:      out,
	}, nil
}

func (m *JWTManager) Generate(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token.Header["kid"] = m.activeKID
	secret, ok := m.keys[m.activeKID]
	if !ok {
		return "", ErrInvalidToken
	}
	return token.SignedString(secret)
}

func (m *JWTManager) Parse(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		kidRaw, ok := token.Header["kid"]
		if !ok {
			return nil, ErrInvalidToken
		}
		kid, ok := kidRaw.(string)
		if !ok || strings.TrimSpace(kid) == "" {
			return nil, ErrInvalidToken
		}
		secret, ok := m.keys[kid]
		if !ok {
			return nil, ErrInvalidToken
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		return 0, ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, ErrInvalidToken
	}
	sub, err := claims.GetSubject()
	if err != nil {
		return 0, ErrInvalidToken
	}
	var id uint
	if _, err := fmt.Sscanf(sub, "%d", &id); err != nil || id == 0 {
		return 0, ErrInvalidToken
	}
	return id, nil
}
