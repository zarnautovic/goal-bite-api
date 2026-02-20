package mealitem

import "time"

type MealItem struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	MealID         uint      `json:"meal_id" gorm:"column:meal_id"`
	FoodID         *uint     `json:"food_id,omitempty" gorm:"column:food_id"`
	RecipeID       *uint     `json:"recipe_id,omitempty" gorm:"column:recipe_id"`
	WeightG        float64   `json:"weight_g" gorm:"column:weight_g"`
	KcalPer100g    float64   `json:"kcal_per_100g" gorm:"column:kcal_per_100g"`
	ProteinPer100g float64   `json:"protein_per_100g" gorm:"column:protein_per_100g"`
	CarbsPer100g   float64   `json:"carbs_per_100g" gorm:"column:carbs_per_100g"`
	FatPer100g     float64   `json:"fat_per_100g" gorm:"column:fat_per_100g"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
