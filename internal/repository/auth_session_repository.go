package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type AuthSession struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"column:user_id"`
	TokenHash string     `gorm:"column:token_hash"`
	ExpiresAt time.Time  `gorm:"column:expires_at"`
	RevokedAt *time.Time `gorm:"column:revoked_at"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at"`
}

func (AuthSession) TableName() string {
	return "auth_sessions"
}

type AuthSessionRepository struct {
	db *gorm.DB
}

func NewAuthSessionRepository(database *gorm.DB) *AuthSessionRepository {
	return &AuthSessionRepository{db: database}
}

type CreateAuthSessionInput struct {
	UserID    uint
	TokenHash string
	ExpiresAt time.Time
}

type RotateAuthSessionInput struct {
	CurrentTokenHash string
	NewTokenHash     string
	UserID           uint
	Now              time.Time
	ExpiresAt        time.Time
}

func (r *AuthSessionRepository) Create(ctx context.Context, in CreateAuthSessionInput) (AuthSession, error) {
	value := AuthSession{
		UserID:    in.UserID,
		TokenHash: in.TokenHash,
		ExpiresAt: in.ExpiresAt,
	}
	if err := r.db.WithContext(ctx).Create(&value).Error; err != nil {
		return AuthSession{}, err
	}
	return value, nil
}

func (r *AuthSessionRepository) GetActiveByTokenHash(ctx context.Context, tokenHash string, now time.Time) (AuthSession, error) {
	var value AuthSession
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		Where("revoked_at IS NULL").
		Where("expires_at > ?", now).
		First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthSession{}, ErrNotFound
	}
	if err != nil {
		return AuthSession{}, err
	}
	return value, nil
}

func (r *AuthSessionRepository) RevokeByID(ctx context.Context, id uint, at time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&AuthSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", at.UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AuthSessionRepository) RevokeByTokenHash(ctx context.Context, tokenHash string, at time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&AuthSession{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", at.UTC())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AuthSessionRepository) Rotate(ctx context.Context, in RotateAuthSessionInput) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		revoke := tx.Model(&AuthSession{}).
			Where("token_hash = ? AND user_id = ? AND revoked_at IS NULL AND expires_at > ?", in.CurrentTokenHash, in.UserID, in.Now.UTC()).
			Update("revoked_at", in.Now.UTC())
		if revoke.Error != nil {
			return revoke.Error
		}
		if revoke.RowsAffected == 0 {
			return ErrNotFound
		}

		value := AuthSession{
			UserID:    in.UserID,
			TokenHash: in.NewTokenHash,
			ExpiresAt: in.ExpiresAt.UTC(),
		}
		if err := tx.Create(&value).Error; err != nil {
			return err
		}
		return nil
	})
}
