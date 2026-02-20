//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestUserGoalsAndProgressE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	upsertUserGoals(t, env.BaseURL, env.Token, 2200, 150, 220, 70)
	foodID := createFood(t, env.BaseURL, env.Token, "Chicken Breast", 165, 31, 0, 3.6)
	createMealWithFoodItem(t, env.BaseURL, foodID, env.Token)

	var goalsOut struct {
		TargetKcal     float64 `json:"target_kcal"`
		TargetProteinG float64 `json:"target_protein_g"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/user-goals", env.BaseURL), nil, env.Token, http.StatusOK, &goalsOut)
	if goalsOut.TargetKcal != 2200 {
		t.Fatalf("expected target kcal 2200, got %v", goalsOut.TargetKcal)
	}

	var progressOut struct {
		Date              string  `json:"date"`
		TotalKcal         float64 `json:"total_kcal"`
		TargetKcal        float64 `json:"target_kcal"`
		RemainingKcal     float64 `json:"remaining_kcal"`
		RemainingProteinG float64 `json:"remaining_protein_g"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/progress/daily?date=2026-02-17", env.BaseURL), nil, env.Token, http.StatusOK, &progressOut)
	if progressOut.Date != "2026-02-17" {
		t.Fatalf("expected date 2026-02-17, got %s", progressOut.Date)
	}
	if progressOut.TargetKcal != 2200 {
		t.Fatalf("expected target kcal 2200, got %v", progressOut.TargetKcal)
	}
	if progressOut.TotalKcal <= 0 {
		t.Fatalf("expected total kcal > 0, got %v", progressOut.TotalKcal)
	}
	if progressOut.RemainingKcal >= progressOut.TargetKcal {
		t.Fatalf("expected remaining kcal to reflect consumed meal, got %v", progressOut.RemainingKcal)
	}
	if progressOut.RemainingProteinG <= 0 {
		t.Fatalf("expected remaining protein > 0, got %v", progressOut.RemainingProteinG)
	}
}
