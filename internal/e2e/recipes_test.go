//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestRecipesE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	foodID := createFood(t, env.BaseURL, env.Token, "Rice", 130, 2.7, 28, 0.3)
	recipeID := createRecipe(t, env.BaseURL, env.Token, foodID)

	var getOut struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/recipes/%d", env.BaseURL, recipeID), nil, env.Token, http.StatusOK, &getOut)
	if getOut.ID != recipeID {
		t.Fatalf("expected recipe id %d, got %d", recipeID, getOut.ID)
	}

	var listOut []struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodGet, env.BaseURL+"/api/v1/recipes?limit=20&offset=0", nil, env.Token, http.StatusOK, &listOut)
	if len(listOut) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(listOut))
	}

	payload := map[string]any{
		"name": "Updated Rice Bowl",
	}
	var updated struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	doJSONWithToken(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/recipes/%d", env.BaseURL, recipeID), payload, env.Token, http.StatusOK, &updated)
	if updated.Name != "Updated Rice Bowl" {
		t.Fatalf("expected updated name Updated Rice Bowl, got %q", updated.Name)
	}
}
