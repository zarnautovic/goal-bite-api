package dto_test

import (
	"errors"
	"testing"

	"nutrition/internal/http/dto"
)

func TestCreateFoodRequestValidate(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		req := dto.CreateFoodRequest{Name: "   ", KcalPer100g: 100, ProteinPer100g: 10, CarbsPer100g: 10, FatPer100g: 10}
		err := req.Validate()
		if !errors.Is(err, dto.ErrInvalidName) {
			t.Fatalf("expected ErrInvalidName, got %v", err)
		}
	})

	t.Run("invalid nutrition", func(t *testing.T) {
		req := dto.CreateFoodRequest{Name: "Rice", KcalPer100g: -1, ProteinPer100g: 1, CarbsPer100g: 1, FatPer100g: 1}
		err := req.Validate()
		if !errors.Is(err, dto.ErrInvalidNutrition) {
			t.Fatalf("expected ErrInvalidNutrition, got %v", err)
		}
	})

	t.Run("valid", func(t *testing.T) {
		req := dto.CreateFoodRequest{Name: "Rice", KcalPer100g: 130, ProteinPer100g: 2.7, CarbsPer100g: 28, FatPer100g: 0.3}
		if err := req.Validate(); err != nil {
			t.Fatalf("expected valid request, got %v", err)
		}
	})
}

func TestUpdateFoodRequestValidate(t *testing.T) {
	t.Run("no fields", func(t *testing.T) {
		req := dto.UpdateFoodRequest{}
		err := req.Validate()
		if !errors.Is(err, dto.ErrNoFieldsToUpdate) {
			t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
		}
	})

	t.Run("invalid name", func(t *testing.T) {
		name := "   "
		req := dto.UpdateFoodRequest{Name: &name}
		err := req.Validate()
		if !errors.Is(err, dto.ErrInvalidName) {
			t.Fatalf("expected ErrInvalidName, got %v", err)
		}
	})

	t.Run("invalid nutrition", func(t *testing.T) {
		kcal := -5.0
		req := dto.UpdateFoodRequest{KcalPer100g: &kcal}
		err := req.Validate()
		if !errors.Is(err, dto.ErrInvalidNutrition) {
			t.Fatalf("expected ErrInvalidNutrition, got %v", err)
		}
	})
}
