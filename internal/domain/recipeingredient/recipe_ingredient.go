package recipeingredient

import "time"

type RecipeIngredient struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	RecipeID   uint      `json:"recipe_id" gorm:"column:recipe_id"`
	FoodID     uint      `json:"food_id" gorm:"column:food_id"`
	RawWeightG float64   `json:"raw_weight_g" gorm:"column:raw_weight_g"`
	Position   *int      `json:"position,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
