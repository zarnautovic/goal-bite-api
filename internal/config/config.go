package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                 string
	Port                   string
	DatabaseURL            string
	MigrationsPath         string
	JWTSecret              string
	JWTActiveKID           string
	JWTKeys                map[string]string
	AuthLoginMaxAttempts   int
	AuthLoginAttemptWindow time.Duration
	AuthLoginLockoutWindow time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                 getEnv("APP_ENV", "development"),
		Port:                   getEnv("PORT", "8080"),
		DatabaseURL:            os.Getenv("DATABASE_URL"),
		MigrationsPath:         getEnv("MIGRATIONS_PATH", "file://internal/db/migrations"),
		JWTSecret:              getEnv("JWT_SECRET", "change-me-dev-secret"),
		JWTActiveKID:           getEnv("JWT_ACTIVE_KID", "v1"),
		JWTKeys:                parseJWTKeys(getEnv("JWT_KEYS", "")),
		AuthLoginMaxAttempts:   getEnvInt("AUTH_LOGIN_MAX_ATTEMPTS", 5),
		AuthLoginAttemptWindow: time.Duration(getEnvInt("AUTH_LOGIN_WINDOW_MINUTES", 10)) * time.Minute,
		AuthLoginLockoutWindow: time.Duration(getEnvInt("AUTH_LOGIN_LOCKOUT_MINUTES", 15)) * time.Minute,
	}
	if len(cfg.JWTKeys) == 0 {
		cfg.JWTKeys = map[string]string{
			cfg.JWTActiveKID: cfg.JWTSecret,
		}
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.Port == "" {
		return Config{}, errors.New("PORT cannot be empty")
	}
	if cfg.MigrationsPath == "" {
		return Config{}, errors.New("MIGRATIONS_PATH cannot be empty")
	}
	if cfg.JWTActiveKID == "" {
		return Config{}, errors.New("JWT_ACTIVE_KID cannot be empty")
	}
	if len(cfg.JWTKeys) == 0 {
		return Config{}, errors.New("JWT key set cannot be empty")
	}
	if _, ok := cfg.JWTKeys[cfg.JWTActiveKID]; !ok {
		return Config{}, errors.New("JWT_ACTIVE_KID must exist in JWT_KEYS")
	}
	if cfg.AuthLoginMaxAttempts <= 0 {
		return Config{}, errors.New("AUTH_LOGIN_MAX_ATTEMPTS must be > 0")
	}
	if cfg.AuthLoginAttemptWindow <= 0 {
		return Config{}, errors.New("AUTH_LOGIN_WINDOW_MINUTES must be > 0")
	}
	if cfg.AuthLoginLockoutWindow <= 0 {
		return Config{}, errors.New("AUTH_LOGIN_LOCKOUT_MINUTES must be > 0")
	}
	return cfg, nil
}

func (c Config) Addr() string {
	return fmt.Sprintf(":%s", c.Port)
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return n
}

func parseJWTKeys(raw string) map[string]string {
	out := make(map[string]string)
	if strings.TrimSpace(raw) == "" {
		return out
	}
	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(kv) != 2 {
			continue
		}
		kid := strings.TrimSpace(kv[0])
		secret := strings.TrimSpace(kv[1])
		if kid == "" || secret == "" {
			continue
		}
		out[kid] = secret
	}
	return out
}
