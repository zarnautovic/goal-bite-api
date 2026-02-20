package dto_test

import (
	"errors"
	"testing"

	"nutrition/internal/http/dto"
)

func TestCreateRecipeRequestValidate(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		req := dto.CreateRecipeRequest{Name: " ", YieldWeightG: 1000, Ingredients: []dto.RecipeIngredientRequest{{FoodID: 1, RawWeightG: 100}}}
		if err := req.Validate(); !errors.Is(err, dto.ErrInvalidRecipeName) {
			t.Fatalf("expected ErrInvalidRecipeName, got %v", err)
		}
	})

	t.Run("invalid ingredients", func(t *testing.T) {
		req := dto.CreateRecipeRequest{Name: "Goulash", YieldWeightG: 1000, Ingredients: []dto.RecipeIngredientRequest{{FoodID: 0, RawWeightG: 100}}}
		if err := req.Validate(); !errors.Is(err, dto.ErrInvalidRecipeIngredients) {
			t.Fatalf("expected ErrInvalidRecipeIngredients, got %v", err)
		}
	})
}

func TestUpdateRecipeRequestValidate(t *testing.T) {
	t.Run("no fields", func(t *testing.T) {
		req := dto.UpdateRecipeRequest{}
		if err := req.Validate(); !errors.Is(err, dto.ErrNoFieldsToUpdate) {
			t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
		}
	})

	t.Run("invalid yield", func(t *testing.T) {
		y := 0.0
		req := dto.UpdateRecipeRequest{YieldWeightG: &y}
		if err := req.Validate(); !errors.Is(err, dto.ErrInvalidRecipeYieldWeight) {
			t.Fatalf("expected ErrInvalidRecipeYieldWeight, got %v", err)
		}
	})
}
