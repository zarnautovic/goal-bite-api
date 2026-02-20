//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestMealsE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	foodID := createFood(t, env.BaseURL, env.Token, "Rice", 130, 2.7, 28, 0.3)
	recipeID := createRecipe(t, env.BaseURL, env.Token, foodID)
	mealID := createMealWithFoodItem(t, env.BaseURL, foodID, env.Token)
	addMealItemRecipe(t, env.BaseURL, mealID, recipeID, env.Token)

	var mealOut struct {
		ID        uint    `json:"id"`
		TotalKcal float64 `json:"total_kcal"`
		Items     []struct {
			ID uint `json:"id"`
		} `json:"items"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/meals/%d", env.BaseURL, mealID), nil, env.Token, http.StatusOK, &mealOut)
	if mealOut.ID != mealID {
		t.Fatalf("expected meal id %d, got %d", mealID, mealOut.ID)
	}
	if len(mealOut.Items) != 2 {
		t.Fatalf("expected 2 meal items, got %d", len(mealOut.Items))
	}
	if mealOut.TotalKcal <= 0 {
		t.Fatalf("expected total kcal > 0, got %v", mealOut.TotalKcal)
	}
	if len(mealOut.Items) < 1 {
		t.Fatalf("expected at least 1 meal item, got %d", len(mealOut.Items))
	}

	firstItemID := mealOut.Items[0].ID
	updateMealPayload := map[string]any{"meal_type": "dinner"}
	doJSONWithToken(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/meals/%d", env.BaseURL, mealID), updateMealPayload, env.Token, http.StatusOK, nil)

	updateItemPayload := map[string]any{"weight_g": 90.0}
	doJSONWithToken(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/meals/%d/items/%d", env.BaseURL, mealID, firstItemID), updateItemPayload, env.Token, http.StatusOK, nil)

	var listOut []struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/meals?date=2026-02-17&limit=20&offset=0", env.BaseURL), nil, env.Token, http.StatusOK, &listOut)
	if len(listOut) != 1 {
		t.Fatalf("expected 1 meal, got %d", len(listOut))
	}

	var totalsOut struct {
		Date      string  `json:"date"`
		TotalKcal float64 `json:"total_kcal"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/daily-totals?date=2026-02-17", env.BaseURL), nil, env.Token, http.StatusOK, &totalsOut)
	if totalsOut.Date != "2026-02-17" {
		t.Fatalf("expected date 2026-02-17, got %s", totalsOut.Date)
	}
	if totalsOut.TotalKcal <= 0 {
		t.Fatalf("expected total kcal > 0, got %v", totalsOut.TotalKcal)
	}

	doJSONWithToken(t, http.MethodDelete, fmt.Sprintf("%s/api/v1/meals/%d/items/%d", env.BaseURL, mealID, firstItemID), nil, env.Token, http.StatusNoContent, nil)
	doJSONWithToken(t, http.MethodDelete, fmt.Sprintf("%s/api/v1/meals/%d", env.BaseURL, mealID), nil, env.Token, http.StatusNoContent, nil)

	var emptyListOut []struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/meals?date=2026-02-17&limit=20&offset=0", env.BaseURL), nil, env.Token, http.StatusOK, &emptyListOut)
	if len(emptyListOut) != 0 {
		t.Fatalf("expected 0 meals after delete, got %d", len(emptyListOut))
	}
}
