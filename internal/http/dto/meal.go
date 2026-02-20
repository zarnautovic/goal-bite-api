package dto

import (
	"errors"
	"time"

	"goal-bite-api/internal/domain/meal"
	"goal-bite-api/internal/service"
)

var (
	ErrInvalidUserID     = errors.New("invalid user id")
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrInvalidMealType   = errors.New("invalid meal type")
	ErrInvalidEatenAt    = errors.New("invalid eaten_at")
	ErrInvalidDate       = errors.New("invalid date")
	ErrInvalidSourceXOR  = errors.New("invalid source xor")
	ErrInvalidItemWeight = errors.New("invalid item weight")
)

type CreateMealRequest struct {
	// Meal type enum.
	MealType meal.MealType `json:"meal_type" example:"lunch"`
	// Meal time in RFC3339 UTC.
	EatenAt string `json:"eaten_at" example:"2026-02-17T12:00:00Z"`
	// Optional items created atomically with meal.
	Items []AddMealItemRequest `json:"items,omitempty"`
}

func (r *CreateMealRequest) Validate() error {
	switch r.MealType {
	case meal.MealTypeBreakfast, meal.MealTypeLunch, meal.MealTypeDinner, meal.MealTypeSnack:
	default:
		return ErrInvalidMealType
	}
	if _, err := time.Parse(time.RFC3339, r.EatenAt); err != nil {
		return ErrInvalidEatenAt
	}
	for _, item := range r.Items {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (r *CreateMealRequest) ToServiceInput(userID uint) (service.CreateMealInput, error) {
	eatenAt, err := time.Parse(time.RFC3339, r.EatenAt)
	if err != nil {
		return service.CreateMealInput{}, err
	}
	items := make([]service.AddMealItemInput, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, item.ToServiceInput())
	}
	return service.CreateMealInput{
		UserID:   userID,
		MealType: r.MealType,
		EatenAt:  eatenAt.UTC(),
		Items:    items,
	}, nil
}

type ListMealsQuery struct {
	Date   string
	Limit  int
	Offset int
}

func (q *ListMealsQuery) Validate() error {
	if _, err := time.Parse("2006-01-02", q.Date); err != nil {
		return ErrInvalidDate
	}
	if !service.IsValidPagination(q.Limit, q.Offset) {
		return ErrInvalidPagination
	}
	return nil
}

func (q *ListMealsQuery) ToServiceInput(userID uint) service.ListMealsInput {
	return service.ListMealsInput{UserID: userID, Date: q.Date, Limit: q.Limit, Offset: q.Offset}
}

type AddMealItemRequest struct {
	// Food source ID (mutually exclusive with recipe_id).
	FoodID *uint `json:"food_id" example:"1"`
	// Recipe source ID (mutually exclusive with food_id).
	RecipeID *uint `json:"recipe_id" example:"1"`
	// Item consumed weight in grams.
	WeightG float64 `json:"weight_g" example:"150"`
}

func (r *AddMealItemRequest) Validate() error {
	foodSet := r.FoodID != nil
	recipeSet := r.RecipeID != nil
	if foodSet == recipeSet {
		return ErrInvalidSourceXOR
	}
	if r.WeightG <= 0 {
		return ErrInvalidItemWeight
	}
	return nil
}

func (r *AddMealItemRequest) ToServiceInput() service.AddMealItemInput {
	return service.AddMealItemInput{FoodID: r.FoodID, RecipeID: r.RecipeID, WeightG: r.WeightG}
}

type UpdateMealRequest struct {
	// Optional meal type enum.
	MealType *meal.MealType `json:"meal_type,omitempty" example:"dinner"`
	// Optional meal time in RFC3339 UTC.
	EatenAt *string `json:"eaten_at,omitempty" example:"2026-02-17T19:00:00Z"`
}

func (r *UpdateMealRequest) Validate() error {
	if r.MealType == nil && r.EatenAt == nil {
		return service.ErrNoFieldsToUpdate
	}
	if r.MealType != nil {
		switch *r.MealType {
		case meal.MealTypeBreakfast, meal.MealTypeLunch, meal.MealTypeDinner, meal.MealTypeSnack:
		default:
			return ErrInvalidMealType
		}
	}
	if r.EatenAt != nil {
		if _, err := time.Parse(time.RFC3339, *r.EatenAt); err != nil {
			return ErrInvalidEatenAt
		}
	}
	return nil
}

func (r *UpdateMealRequest) ToServiceInput() (service.UpdateMealInput, error) {
	out := service.UpdateMealInput{MealType: r.MealType}
	if r.EatenAt != nil {
		v, err := time.Parse(time.RFC3339, *r.EatenAt)
		if err != nil {
			return service.UpdateMealInput{}, err
		}
		v = v.UTC()
		out.EatenAt = &v
	}
	return out, nil
}

type UpdateMealItemRequest struct {
	// Optional food source ID (mutually exclusive with recipe_id).
	FoodID *uint `json:"food_id,omitempty" example:"1"`
	// Optional recipe source ID (mutually exclusive with food_id).
	RecipeID *uint `json:"recipe_id,omitempty" example:"1"`
	// Optional consumed weight in grams.
	WeightG *float64 `json:"weight_g,omitempty" example:"180"`
}

func (r *UpdateMealItemRequest) Validate() error {
	if r.FoodID == nil && r.RecipeID == nil && r.WeightG == nil {
		return service.ErrNoFieldsToUpdate
	}
	if r.FoodID != nil && r.RecipeID != nil {
		return ErrInvalidSourceXOR
	}
	if r.WeightG != nil && *r.WeightG <= 0 {
		return ErrInvalidItemWeight
	}
	return nil
}

func (r *UpdateMealItemRequest) ToServiceInput() service.UpdateMealItemInput {
	return service.UpdateMealItemInput{
		FoodID:   r.FoodID,
		RecipeID: r.RecipeID,
		WeightG:  r.WeightG,
	}
}
