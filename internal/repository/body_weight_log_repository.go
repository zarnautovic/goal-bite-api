package repository

import (
	"context"
	"errors"
	"time"

	"nutrition/internal/domain/bodyweightlog"

	"gorm.io/gorm"
)

type BodyWeightLogRepository struct {
	db *gorm.DB
}

type CreateBodyWeightLogInput struct {
	UserID   uint
	WeightKG float64
	LoggedAt time.Time
}

func NewBodyWeightLogRepository(database *gorm.DB) *BodyWeightLogRepository {
	return &BodyWeightLogRepository{db: database}
}

func (r *BodyWeightLogRepository) Create(ctx context.Context, in CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error) {
	value := bodyweightlog.BodyWeightLog{
		UserID:   in.UserID,
		WeightKG: in.WeightKG,
		LoggedAt: in.LoggedAt,
	}
	if err := r.db.WithContext(ctx).Create(&value).Error; err != nil {
		return bodyweightlog.BodyWeightLog{}, err
	}
	return value, nil
}

func (r *BodyWeightLogRepository) ListByRange(ctx context.Context, userID uint, from, to time.Time, limit, offset int) ([]bodyweightlog.BodyWeightLog, error) {
	var out []bodyweightlog.BodyWeightLog
	err := r.db.WithContext(ctx).
		Select("id, user_id, weight_kg, logged_at, created_at, updated_at").
		Where("user_id = ?", userID).
		Where("logged_at >= ? AND logged_at < ?", from, to).
		Order("logged_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *BodyWeightLogRepository) GetLatest(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error) {
	var out bodyweightlog.BodyWeightLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("logged_at DESC, id DESC").
		First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return bodyweightlog.BodyWeightLog{}, ErrNotFound
	}
	if err != nil {
		return bodyweightlog.BodyWeightLog{}, err
	}
	return out, nil
}

func (r *BodyWeightLogRepository) ListByRangeAll(ctx context.Context, userID uint, from, to time.Time) ([]bodyweightlog.BodyWeightLog, error) {
	var out []bodyweightlog.BodyWeightLog
	err := r.db.WithContext(ctx).
		Select("id, user_id, weight_kg, logged_at, created_at, updated_at").
		Where("user_id = ?", userID).
		Where("logged_at >= ? AND logged_at < ?", from, to).
		Order("logged_at ASC, id ASC").
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}
