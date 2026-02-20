package meal

import (
	"time"

	"nutrition/internal/domain/mealitem"
)

type MealType string

const (
	MealTypeBreakfast MealType = "breakfast"
	MealTypeLunch     MealType = "lunch"
	MealTypeDinner    MealType = "dinner"
	MealTypeSnack     MealType = "snack"
)

type Meal struct {
	ID        uint                `json:"id" gorm:"primaryKey"`
	UserID    uint                `json:"user_id" gorm:"column:user_id"`
	MealType  MealType            `json:"meal_type" gorm:"column:meal_type"`
	EatenAt   time.Time           `json:"eaten_at" gorm:"column:eaten_at"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Items     []mealitem.MealItem `json:"items,omitempty" gorm:"-"`
}
