package repository

import (
	"context"
	"errors"

	"nutrition/internal/domain/usergoal"

	"gorm.io/gorm"
)

type UserGoalRepository struct {
	db *gorm.DB
}

type UpsertUserGoalInput struct {
	UserID         uint
	TargetKcal     float64
	TargetProteinG float64
	TargetCarbsG   float64
	TargetFatG     float64
	WeightGoalKG   *float64
	ActivityLevel  *string
}

func NewUserGoalRepository(database *gorm.DB) *UserGoalRepository {
	return &UserGoalRepository{db: database}
}

func (r *UserGoalRepository) Upsert(ctx context.Context, in UpsertUserGoalInput) (usergoal.UserGoal, error) {
	var existing usergoal.UserGoal
	err := r.db.WithContext(ctx).Where("user_id = ?", in.UserID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		value := usergoal.UserGoal{
			UserID:         in.UserID,
			TargetKcal:     in.TargetKcal,
			TargetProteinG: in.TargetProteinG,
			TargetCarbsG:   in.TargetCarbsG,
			TargetFatG:     in.TargetFatG,
			WeightGoalKG:   in.WeightGoalKG,
			ActivityLevel:  in.ActivityLevel,
		}
		if createErr := r.db.WithContext(ctx).Create(&value).Error; createErr != nil {
			return usergoal.UserGoal{}, createErr
		}
		return value, nil
	}
	if err != nil {
		return usergoal.UserGoal{}, err
	}

	existing.TargetKcal = in.TargetKcal
	existing.TargetProteinG = in.TargetProteinG
	existing.TargetCarbsG = in.TargetCarbsG
	existing.TargetFatG = in.TargetFatG
	existing.WeightGoalKG = in.WeightGoalKG
	existing.ActivityLevel = in.ActivityLevel

	if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
		return usergoal.UserGoal{}, err
	}
	return existing, nil
}

func (r *UserGoalRepository) GetByUserID(ctx context.Context, userID uint) (usergoal.UserGoal, error) {
	var value usergoal.UserGoal
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&value).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return usergoal.UserGoal{}, ErrNotFound
	}
	if err != nil {
		return usergoal.UserGoal{}, err
	}
	return value, nil
}
