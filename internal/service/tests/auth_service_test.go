package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goal-bite-api/internal/auth"
	"goal-bite-api/internal/domain/user"
	"goal-bite-api/internal/repository"
	"goal-bite-api/internal/service"
)

type fakeUserAuthStore struct {
	getByIDFn           func(ctx context.Context, id uint) (user.User, error)
	getByEmailFn        func(ctx context.Context, email string) (user.User, error)
	createFn            func(ctx context.Context, value user.User) (user.User, error)
	createWithSessionFn func(ctx context.Context, value user.User, session repository.CreateAuthSessionInput) (user.User, error)
}

func (f fakeUserAuthStore) GetByID(ctx context.Context, id uint) (user.User, error) {
	if f.getByIDFn == nil {
		return user.User{}, nil
	}
	return f.getByIDFn(ctx, id)
}

func (f fakeUserAuthStore) GetByEmail(ctx context.Context, email string) (user.User, error) {
	if f.getByEmailFn == nil {
		return user.User{}, repository.ErrNotFound
	}
	return f.getByEmailFn(ctx, email)
}

func (f fakeUserAuthStore) Create(ctx context.Context, value user.User) (user.User, error) {
	if f.createFn == nil {
		value.ID = 1
		return value, nil
	}
	return f.createFn(ctx, value)
}

func (f fakeUserAuthStore) CreateWithSession(ctx context.Context, value user.User, session repository.CreateAuthSessionInput) (user.User, error) {
	if f.createWithSessionFn == nil {
		if f.createFn != nil {
			return f.createFn(ctx, value)
		}
		value.ID = 1
		return value, nil
	}
	return f.createWithSessionFn(ctx, value, session)
}

type fakeTokenIssuer struct {
	generateFn func(userID uint) (string, error)
}

func (f fakeTokenIssuer) Generate(userID uint) (string, error) {
	if f.generateFn == nil {
		return "access-token", nil
	}
	return f.generateFn(userID)
}

type fakeAuthSessionStore struct {
	createFn            func(ctx context.Context, in repository.CreateAuthSessionInput) (repository.AuthSession, error)
	getActiveByHashFn   func(ctx context.Context, tokenHash string, now time.Time) (repository.AuthSession, error)
	revokeByIDFn        func(ctx context.Context, id uint, at time.Time) error
	revokeByTokenHashFn func(ctx context.Context, tokenHash string, at time.Time) error
	rotateFn            func(ctx context.Context, in repository.RotateAuthSessionInput) error
}

type fakeLoginAttemptTracker struct {
	isBlockedFn       func(key string, now time.Time) (bool, time.Duration)
	registerFailureFn func(key string, now time.Time)
	resetFn           func(key string)
}

func (f fakeLoginAttemptTracker) IsBlocked(key string, now time.Time) (bool, time.Duration) {
	if f.isBlockedFn == nil {
		return false, 0
	}
	return f.isBlockedFn(key, now)
}

func (f fakeLoginAttemptTracker) RegisterFailure(key string, now time.Time) {
	if f.registerFailureFn != nil {
		f.registerFailureFn(key, now)
	}
}

func (f fakeLoginAttemptTracker) Reset(key string) {
	if f.resetFn != nil {
		f.resetFn(key)
	}
}

func (f fakeAuthSessionStore) Create(ctx context.Context, in repository.CreateAuthSessionInput) (repository.AuthSession, error) {
	if f.createFn == nil {
		return repository.AuthSession{ID: 1, UserID: in.UserID}, nil
	}
	return f.createFn(ctx, in)
}

func (f fakeAuthSessionStore) GetActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (repository.AuthSession, error) {
	if f.getActiveByHashFn == nil {
		return repository.AuthSession{}, repository.ErrNotFound
	}
	return f.getActiveByHashFn(ctx, tokenHash, now)
}

func (f fakeAuthSessionStore) RevokeByID(ctx context.Context, id uint, at time.Time) error {
	if f.revokeByIDFn == nil {
		return nil
	}
	return f.revokeByIDFn(ctx, id, at)
}

func (f fakeAuthSessionStore) RevokeByTokenHash(ctx context.Context, tokenHash string, at time.Time) error {
	if f.revokeByTokenHashFn == nil {
		return nil
	}
	return f.revokeByTokenHashFn(ctx, tokenHash, at)
}

func (f fakeAuthSessionStore) Rotate(ctx context.Context, in repository.RotateAuthSessionInput) error {
	if f.rotateFn == nil {
		return nil
	}
	return f.rotateFn(ctx, in)
}

func TestAuthServiceRefresh(t *testing.T) {
	t.Run("invalid token maps to invalid refresh token", func(t *testing.T) {
		svc := service.NewAuthService(fakeUserAuthStore{}, fakeTokenIssuer{}, fakeAuthSessionStore{})
		_, err := svc.Refresh(context.Background(), "bad")
		if !errors.Is(err, service.ErrInvalidRefreshToken) {
			t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("refresh rotates tokens", func(t *testing.T) {
		rotated := false
		svc := service.NewAuthService(
			fakeUserAuthStore{
				getByIDFn: func(_ context.Context, id uint) (user.User, error) {
					return user.User{ID: id, Name: "User", Email: "u@example.com"}, nil
				},
			},
			fakeTokenIssuer{generateFn: func(_ uint) (string, error) { return "new-access", nil }},
			fakeAuthSessionStore{
				getActiveByHashFn: func(_ context.Context, _ string, _ time.Time) (repository.AuthSession, error) {
					return repository.AuthSession{ID: 7, UserID: 1}, nil
				},
				rotateFn: func(_ context.Context, in repository.RotateAuthSessionInput) error {
					if in.UserID != 1 {
						t.Fatalf("expected user id 1, got %d", in.UserID)
					}
					if in.CurrentTokenHash == "" || in.NewTokenHash == "" {
						t.Fatalf("expected token hashes to be set")
					}
					rotated = true
					return nil
				},
			},
		)

		out, err := svc.Refresh(context.Background(), "valid-refresh")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.AccessToken != "new-access" || out.Token != "new-access" {
			t.Fatalf("unexpected access token result: %+v", out)
		}
		if out.RefreshToken == "" {
			t.Fatalf("expected rotated refresh token")
		}
		if !rotated {
			t.Fatalf("expected rotate to be called")
		}
	})
}

func TestAuthServiceLogout(t *testing.T) {
	t.Run("invalid token maps to invalid refresh token", func(t *testing.T) {
		svc := service.NewAuthService(fakeUserAuthStore{}, fakeTokenIssuer{}, fakeAuthSessionStore{
			revokeByTokenHashFn: func(_ context.Context, _ string, _ time.Time) error {
				return repository.ErrNotFound
			},
		})
		err := svc.Logout(context.Background(), "bad")
		if !errors.Is(err, service.ErrInvalidRefreshToken) {
			t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
		}
	})

	t.Run("valid token revokes session", func(t *testing.T) {
		revoked := false
		svc := service.NewAuthService(fakeUserAuthStore{}, fakeTokenIssuer{}, fakeAuthSessionStore{
			revokeByTokenHashFn: func(_ context.Context, tokenHash string, _ time.Time) error {
				if tokenHash == "" {
					t.Fatalf("expected hashed token")
				}
				revoked = true
				return nil
			},
		})
		err := svc.Logout(context.Background(), "valid-refresh-token")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !revoked {
			t.Fatalf("expected session revocation call")
		}
	})
}

func TestAuthServiceRegisterInvalidProfile(t *testing.T) {
	t.Run("invalid sex returns invalid profile", func(t *testing.T) {
		invalid := "unknown"
		svc := service.NewAuthService(fakeUserAuthStore{}, fakeTokenIssuer{}, fakeAuthSessionStore{})
		_, err := svc.Register(context.Background(), service.RegisterInput{
			Name:     "A",
			Email:    "a@example.com",
			Password: "SuperSecret1!",
			Sex:      &invalid,
		})
		if !errors.Is(err, service.ErrInvalidProfile) {
			t.Fatalf("expected ErrInvalidProfile, got %v", err)
		}
	})
}

func TestAuthServiceRegisterUsesCreateWithSession(t *testing.T) {
	called := false
	svc := service.NewAuthService(
		fakeUserAuthStore{
			createWithSessionFn: func(_ context.Context, value user.User, session repository.CreateAuthSessionInput) (user.User, error) {
				if session.TokenHash == "" {
					t.Fatalf("expected session token hash")
				}
				if session.ExpiresAt.IsZero() {
					t.Fatalf("expected session expiry")
				}
				called = true
				value.ID = 1
				return value, nil
			},
		},
		fakeTokenIssuer{},
		fakeAuthSessionStore{},
	)
	_, err := svc.Register(context.Background(), service.RegisterInput{
		Name:     "A",
		Email:    "a@example.com",
		Password: "SuperSecret1!",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !called {
		t.Fatalf("expected CreateWithSession to be called")
	}
}

func TestAuthServiceLoginAttemptLockout(t *testing.T) {
	t.Run("blocked login returns too many attempts", func(t *testing.T) {
		svc := service.NewAuthService(
			fakeUserAuthStore{},
			fakeTokenIssuer{},
			fakeAuthSessionStore{},
			fakeLoginAttemptTracker{
				isBlockedFn: func(_ string, _ time.Time) (bool, time.Duration) { return true, time.Minute },
			},
		)
		_, err := svc.Login(context.Background(), "a@example.com", "Pass1234!")
		if !errors.Is(err, service.ErrTooManyLoginAttempts) {
			t.Fatalf("expected ErrTooManyLoginAttempts, got %v", err)
		}
	})

	t.Run("successful login resets attempts", func(t *testing.T) {
		hash, err := auth.HashPassword("password")
		if err != nil {
			t.Fatalf("hash password: %v", err)
		}
		resetCalled := false
		svc := service.NewAuthService(
			fakeUserAuthStore{
				getByEmailFn: func(_ context.Context, _ string) (user.User, error) {
					return user.User{
						ID:           1,
						Email:        "a@example.com",
						Name:         "A",
						PasswordHash: hash,
					}, nil
				},
			},
			fakeTokenIssuer{},
			fakeAuthSessionStore{},
			fakeLoginAttemptTracker{
				isBlockedFn: func(_ string, _ time.Time) (bool, time.Duration) { return false, 0 },
				resetFn:     func(_ string) { resetCalled = true },
			},
		)
		if _, err := svc.Login(context.Background(), "a@example.com", "password"); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !resetCalled {
			t.Fatalf("expected reset to be called")
		}
	})
}
