package handlers

import (
	"nutrition/internal/domain/meal"
	"nutrition/internal/domain/mealitem"
)

type mealResponse struct {
	meal.Meal
	TotalKcal     float64 `json:"total_kcal"`
	TotalProteinG float64 `json:"total_protein_g"`
	TotalCarbsG   float64 `json:"total_carbs_g"`
	TotalFatG     float64 `json:"total_fat_g"`
}

func toMealResponse(value meal.Meal) mealResponse {
	totalKcal, totalProtein, totalCarbs, totalFat := calculateMealTotals(value.Items)
	return mealResponse{
		Meal:          value,
		TotalKcal:     totalKcal,
		TotalProteinG: totalProtein,
		TotalCarbsG:   totalCarbs,
		TotalFatG:     totalFat,
	}
}

func toMealResponses(values []meal.Meal) []mealResponse {
	out := make([]mealResponse, 0, len(values))
	for _, value := range values {
		out = append(out, toMealResponse(value))
	}
	return out
}

func calculateMealTotals(items []mealitem.MealItem) (float64, float64, float64, float64) {
	var kcal float64
	var protein float64
	var carbs float64
	var fat float64

	for _, item := range items {
		ratio := item.WeightG / 100.0
		kcal += item.KcalPer100g * ratio
		protein += item.ProteinPer100g * ratio
		carbs += item.CarbsPer100g * ratio
		fat += item.FatPer100g * ratio
	}

	return kcal, protein, carbs, fat
}
