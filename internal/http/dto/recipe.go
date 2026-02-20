package dto

import (
	"errors"
	"strings"

	"nutrition/internal/service"
)

var (
	ErrInvalidRecipeName        = errors.New("invalid recipe name")
	ErrInvalidRecipeYieldWeight = errors.New("invalid recipe yield weight")
	ErrInvalidRecipeIngredients = errors.New("invalid recipe ingredients")
)

type RecipeIngredientRequest struct {
	// Existing food ID used as ingredient source.
	FoodID uint `json:"food_id" example:"1"`
	// Raw ingredient weight in grams.
	RawWeightG float64 `json:"raw_weight_g" example:"200"`
	// Optional ordering position.
	Position *int `json:"position,omitempty" example:"1"`
}

type CreateRecipeRequest struct {
	// Human-readable recipe name.
	Name string `json:"name" example:"Rice Bowl"`
	// Final cooked yield weight in grams.
	YieldWeightG float64 `json:"yield_weight_g" example:"200"`
	// Ingredient list used for nutrition calculation.
	Ingredients []RecipeIngredientRequest `json:"ingredients"`
}

func (r *CreateRecipeRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrInvalidRecipeName
	}
	if r.YieldWeightG <= 0 {
		return ErrInvalidRecipeYieldWeight
	}
	if len(r.Ingredients) == 0 {
		return ErrInvalidRecipeIngredients
	}
	for _, item := range r.Ingredients {
		if item.FoodID == 0 || item.RawWeightG <= 0 {
			return ErrInvalidRecipeIngredients
		}
	}
	return nil
}

func (r *CreateRecipeRequest) ToServiceInput() service.CreateRecipeInput {
	return service.CreateRecipeInput{
		Name:         r.Name,
		YieldWeightG: r.YieldWeightG,
		Ingredients:  toServiceRecipeIngredients(r.Ingredients),
	}
}

type UpdateRecipeRequest struct {
	// Optional recipe name.
	Name *string `json:"name" example:"Updated Rice Bowl"`
	// Optional final cooked yield weight in grams.
	YieldWeightG *float64 `json:"yield_weight_g" example:"210"`
	// Optional full replacement of ingredient list.
	Ingredients *[]RecipeIngredientRequest `json:"ingredients"`
}

func (r *UpdateRecipeRequest) Validate() error {
	if r.Name == nil && r.YieldWeightG == nil && r.Ingredients == nil {
		return ErrNoFieldsToUpdate
	}
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrInvalidRecipeName
	}
	if r.YieldWeightG != nil && *r.YieldWeightG <= 0 {
		return ErrInvalidRecipeYieldWeight
	}
	if r.Ingredients != nil {
		if len(*r.Ingredients) == 0 {
			return ErrInvalidRecipeIngredients
		}
		for _, item := range *r.Ingredients {
			if item.FoodID == 0 || item.RawWeightG <= 0 {
				return ErrInvalidRecipeIngredients
			}
		}
	}
	return nil
}

func (r *UpdateRecipeRequest) ToServiceInput() service.UpdateRecipeInput {
	input := service.UpdateRecipeInput{
		Name:         r.Name,
		YieldWeightG: r.YieldWeightG,
	}
	if r.Ingredients != nil {
		mapped := toServiceRecipeIngredients(*r.Ingredients)
		input.Ingredients = &mapped
	}
	return input
}

func toServiceRecipeIngredients(items []RecipeIngredientRequest) []service.RecipeIngredientInput {
	out := make([]service.RecipeIngredientInput, 0, len(items))
	for _, item := range items {
		out = append(out, service.RecipeIngredientInput{
			FoodID:     item.FoodID,
			RawWeightG: item.RawWeightG,
			Position:   item.Position,
		})
	}
	return out
}
