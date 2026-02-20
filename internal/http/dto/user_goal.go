package dto

import (
	"errors"

	"goal-bite-api/internal/service"
)

var (
	ErrInvalidTargetKcal     = errors.New("invalid target_kcal")
	ErrInvalidTargetProteinG = errors.New("invalid target_protein_g")
	ErrInvalidTargetCarbsG   = errors.New("invalid target_carbs_g")
	ErrInvalidTargetFatG     = errors.New("invalid target_fat_g")
	ErrInvalidWeightGoalKG   = errors.New("invalid weight_goal_kg")
	ErrInvalidActivityLevel  = errors.New("invalid activity_level")
)

type UpsertUserGoalRequest struct {
	// Daily kcal target.
	TargetKcal float64 `json:"target_kcal" example:"2200"`
	// Daily protein target in grams.
	TargetProteinG float64 `json:"target_protein_g" example:"150"`
	// Daily carbs target in grams.
	TargetCarbsG float64 `json:"target_carbs_g" example:"220"`
	// Daily fat target in grams.
	TargetFatG float64 `json:"target_fat_g" example:"70"`
	// Optional body weight goal in kilograms.
	WeightGoalKG *float64 `json:"weight_goal_kg,omitempty" example:"80"`
	// Optional activity level: sedentary|light|moderate|active|very_active.
	ActivityLevel *string `json:"activity_level,omitempty" example:"moderate"`
}

func (r *UpsertUserGoalRequest) Validate() error {
	if r.TargetKcal <= 0 {
		return ErrInvalidTargetKcal
	}
	if r.TargetProteinG <= 0 {
		return ErrInvalidTargetProteinG
	}
	if r.TargetCarbsG <= 0 {
		return ErrInvalidTargetCarbsG
	}
	if r.TargetFatG <= 0 {
		return ErrInvalidTargetFatG
	}
	if r.WeightGoalKG != nil && *r.WeightGoalKG <= 0 {
		return ErrInvalidWeightGoalKG
	}
	if r.ActivityLevel != nil {
		switch *r.ActivityLevel {
		case "sedentary", "light", "moderate", "active", "very_active":
		default:
			return ErrInvalidActivityLevel
		}
	}
	return nil
}

func (r *UpsertUserGoalRequest) ToServiceInput(userID uint) service.UpsertUserGoalInput {
	return service.UpsertUserGoalInput{
		UserID:         userID,
		TargetKcal:     r.TargetKcal,
		TargetProteinG: r.TargetProteinG,
		TargetCarbsG:   r.TargetCarbsG,
		TargetFatG:     r.TargetFatG,
		WeightGoalKG:   r.WeightGoalKG,
		ActivityLevel:  r.ActivityLevel,
	}
}
