package recipe

import (
	"time"

	"goal-bite-api/internal/domain/recipeingredient"
)

type Recipe struct {
	ID             uint                                `json:"id" gorm:"primaryKey"`
	UserID         uint                                `json:"user_id" gorm:"column:user_id;not null"`
	Name           string                              `json:"name"`
	YieldWeightG   float64                             `json:"yield_weight_g" gorm:"column:yield_weight_g"`
	KcalPer100g    float64                             `json:"kcal_per_100g" gorm:"column:kcal_per_100g"`
	ProteinPer100g float64                             `json:"protein_per_100g" gorm:"column:protein_per_100g"`
	CarbsPer100g   float64                             `json:"carbs_per_100g" gorm:"column:carbs_per_100g"`
	FatPer100g     float64                             `json:"fat_per_100g" gorm:"column:fat_per_100g"`
	CreatedAt      time.Time                           `json:"created_at"`
	UpdatedAt      time.Time                           `json:"updated_at"`
	Ingredients    []recipeingredient.RecipeIngredient `json:"ingredients,omitempty" gorm:"-"`
}
