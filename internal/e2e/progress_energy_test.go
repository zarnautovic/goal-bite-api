//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestEnergyProgressE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	foodID := createFood(t, env.BaseURL, env.Token, "Chicken", 165, 31, 0, 3.6)
	createMealWithFoodItem(t, env.BaseURL, foodID, env.Token)
	createBodyWeightLog(t, env.BaseURL, 85.0, env.Token)

	// Add one more weight log to allow trend calculation.
	payload := map[string]any{
		"weight_kg": 84.8,
		"logged_at": "2026-02-18T08:00:00Z",
	}
	doJSONWithToken(t, http.MethodPost, env.BaseURL+"/api/v1/body-weight-logs", payload, env.Token, http.StatusCreated, nil)

	// Enrich profile for formula branch.
	updateMe := map[string]any{
		"sex":            "male",
		"birth_date":     "1994-05-18",
		"height_cm":      178.0,
		"activity_level": "moderate",
	}
	doJSONWithToken(t, http.MethodPatch, env.BaseURL+"/api/v1/users/me", updateMe, env.Token, http.StatusOK, nil)

	var out struct {
		ObservedTDEEKcal    float64  `json:"observed_tdee_kcal"`
		RecommendedTDEEKcal float64  `json:"recommended_tdee_kcal"`
		FormulaTDEEKcal     *float64 `json:"formula_tdee_kcal"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/progress/energy?from=2026-02-01&to=2026-02-18", env.BaseURL), nil, env.Token, http.StatusOK, &out)
	if out.ObservedTDEEKcal == 0 || out.RecommendedTDEEKcal == 0 {
		t.Fatalf("unexpected output: %+v", out)
	}
	if out.FormulaTDEEKcal == nil {
		t.Fatalf("expected formula_tdee_kcal to be present")
	}
}
