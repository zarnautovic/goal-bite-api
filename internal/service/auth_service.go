package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"goal-bite-api/internal/auth"
	"goal-bite-api/internal/domain/user"
	"goal-bite-api/internal/repository"
)

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrTooManyLoginAttempts = errors.New("too many login attempts")
	ErrInvalidName          = errors.New("invalid name")
	ErrInvalidProfile       = errors.New("invalid profile fields")
	ErrInvalidRefreshToken  = errors.New("invalid refresh token")
)

type UserAuthStore interface {
	GetByID(ctx context.Context, id uint) (user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
	Create(ctx context.Context, value user.User) (user.User, error)
	CreateWithSession(ctx context.Context, value user.User, session repository.CreateAuthSessionInput) (user.User, error)
}

type TokenIssuer interface {
	Generate(userID uint) (string, error)
}

type AuthSessionStore interface {
	Create(ctx context.Context, in repository.CreateAuthSessionInput) (repository.AuthSession, error)
	GetActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (repository.AuthSession, error)
	RevokeByID(ctx context.Context, id uint, at time.Time) error
	RevokeByTokenHash(ctx context.Context, tokenHash string, at time.Time) error
	Rotate(ctx context.Context, in repository.RotateAuthSessionInput) error
}

type AuthService struct {
	users    UserAuthStore
	tokens   TokenIssuer
	sessions AuthSessionStore
	attempts LoginAttemptTracker
}

type RegisterInput struct {
	Name          string
	Email         string
	Password      string
	Sex           *string
	BirthDate     *time.Time
	HeightCM      *float64
	ActivityLevel *string
}

type AuthResult struct {
	Token        string    `json:"token"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         user.User `json:"user"`
}

func NewAuthService(users UserAuthStore, tokens TokenIssuer, sessions AuthSessionStore, attempts ...LoginAttemptTracker) *AuthService {
	tracker := LoginAttemptTracker(NewMemoryLoginAttemptTracker(5, 10*time.Minute, 15*time.Minute))
	if len(attempts) > 0 && attempts[0] != nil {
		tracker = attempts[0]
	}
	return &AuthService{users: users, tokens: tokens, sessions: sessions, attempts: tracker}
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (AuthResult, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return AuthResult{}, ErrInvalidName
	}
	email, err := auth.NormalizeEmail(in.Email)
	if err != nil {
		return AuthResult{}, ErrInvalidEmail
	}
	if !auth.ValidatePasswordPolicy(in.Password) {
		return AuthResult{}, ErrInvalidPassword
	}
	if in.Sex != nil {
		switch *in.Sex {
		case "male", "female":
		default:
			return AuthResult{}, ErrInvalidProfile
		}
	}
	if in.HeightCM != nil && *in.HeightCM <= 0 {
		return AuthResult{}, ErrInvalidProfile
	}
	if in.ActivityLevel != nil {
		switch *in.ActivityLevel {
		case "sedentary", "light", "moderate", "active", "very_active":
		default:
			return AuthResult{}, ErrInvalidProfile
		}
	}

	_, err = s.users.GetByEmail(ctx, email)
	if err == nil {
		return AuthResult{}, ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return AuthResult{}, err
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return AuthResult{}, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	created, err := s.users.CreateWithSession(ctx, user.User{
		Name:          name,
		Email:         email,
		Sex:           in.Sex,
		BirthDate:     in.BirthDate,
		HeightCM:      in.HeightCM,
		ActivityLevel: in.ActivityLevel,
		PasswordHash:  hash,
	}, repository.CreateAuthSessionInput{
		TokenHash: hashRefreshToken(refreshToken),
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		return AuthResult{}, err
	}

	token, err := s.tokens.Generate(created.ID)
	if err != nil {
		return AuthResult{}, err
	}
	created.PasswordHash = ""
	return AuthResult{
		Token:        token,
		AccessToken:  token,
		RefreshToken: refreshToken,
		User:         created,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, emailRaw, password string) (AuthResult, error) {
	email, err := auth.NormalizeEmail(emailRaw)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	now := time.Now().UTC()
	if blocked, _ := s.attempts.IsBlocked(email, now); blocked {
		return AuthResult{}, ErrTooManyLoginAttempts
	}

	u, err := s.users.GetByEmail(ctx, email)
	if errors.Is(err, repository.ErrNotFound) {
		s.attempts.RegisterFailure(email, now)
		if blocked, _ := s.attempts.IsBlocked(email, now); blocked {
			return AuthResult{}, ErrTooManyLoginAttempts
		}
		return AuthResult{}, ErrInvalidCredentials
	}
	if err != nil {
		return AuthResult{}, err
	}

	if !auth.CheckPassword(u.PasswordHash, password) {
		s.attempts.RegisterFailure(email, now)
		if blocked, _ := s.attempts.IsBlocked(email, now); blocked {
			return AuthResult{}, ErrTooManyLoginAttempts
		}
		return AuthResult{}, ErrInvalidCredentials
	}
	s.attempts.Reset(email)

	token, err := s.tokens.Generate(u.ID)
	if err != nil {
		return AuthResult{}, err
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}
	if _, err := s.sessions.Create(ctx, repository.CreateAuthSessionInput{
		UserID:    u.ID,
		TokenHash: hashRefreshToken(refreshToken),
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
	}); err != nil {
		return AuthResult{}, err
	}
	u.PasswordHash = ""
	return AuthResult{
		Token:        token,
		AccessToken:  token,
		RefreshToken: refreshToken,
		User:         u,
	}, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (AuthResult, error) {
	token := strings.TrimSpace(refreshToken)
	if token == "" {
		return AuthResult{}, ErrInvalidRefreshToken
	}

	now := time.Now().UTC()
	session, err := s.sessions.GetActiveByTokenHash(ctx, hashRefreshToken(token), now)
	if errors.Is(err, repository.ErrNotFound) {
		return AuthResult{}, ErrInvalidRefreshToken
	}
	if err != nil {
		return AuthResult{}, err
	}

	u, err := s.users.GetByID(ctx, session.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return AuthResult{}, ErrInvalidRefreshToken
	}
	if err != nil {
		return AuthResult{}, err
	}

	accessToken, err := s.tokens.Generate(session.UserID)
	if err != nil {
		return AuthResult{}, err
	}
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return AuthResult{}, err
	}

	if err := s.sessions.Rotate(ctx, repository.RotateAuthSessionInput{
		CurrentTokenHash: hashRefreshToken(token),
		NewTokenHash:     hashRefreshToken(newRefreshToken),
		UserID:           session.UserID,
		Now:              now,
		ExpiresAt:        now.Add(30 * 24 * time.Hour),
	}); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return AuthResult{}, ErrInvalidRefreshToken
		}
		return AuthResult{}, err
	}

	u.PasswordHash = ""
	return AuthResult{
		Token:        accessToken,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         u,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	token := strings.TrimSpace(refreshToken)
	if token == "" {
		return ErrInvalidRefreshToken
	}
	err := s.sessions.RevokeByTokenHash(ctx, hashRefreshToken(token), time.Now().UTC())
	if errors.Is(err, repository.ErrNotFound) {
		return ErrInvalidRefreshToken
	}
	return err
}

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
