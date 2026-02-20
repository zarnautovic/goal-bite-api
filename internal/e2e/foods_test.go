//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestFoodsE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	foodID := createFood(t, env.BaseURL, env.Token, "Rice", 130, 2.7, 28, 0.3)

	var getOut struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/foods/%d", env.BaseURL, foodID), nil, env.Token, http.StatusOK, &getOut)
	if getOut.ID != foodID {
		t.Fatalf("expected food id %d, got %d", foodID, getOut.ID)
	}

	var listOut []struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodGet, env.BaseURL+"/api/v1/foods?limit=20&offset=0", nil, env.Token, http.StatusOK, &listOut)
	if len(listOut) != 1 {
		t.Fatalf("expected 1 food, got %d", len(listOut))
	}

	payload := map[string]any{
		"name": "Cooked Rice",
	}
	var updated struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}
	doJSONWithToken(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/foods/%d", env.BaseURL, foodID), payload, env.Token, http.StatusOK, &updated)
	if updated.Name != "Cooked Rice" {
		t.Fatalf("expected updated name Cooked Rice, got %q", updated.Name)
	}

	barcodePayload := map[string]any{
		"name":             "Protein Bar",
		"barcode":          "5901234123457",
		"kcal_per_100g":    410.0,
		"protein_per_100g": 30.0,
		"carbs_per_100g":   40.0,
		"fat_per_100g":     12.0,
	}
	var byBarcodeCreated struct {
		ID      uint    `json:"id"`
		Barcode *string `json:"barcode"`
	}
	doJSONWithToken(t, http.MethodPost, env.BaseURL+"/api/v1/foods", barcodePayload, env.Token, http.StatusCreated, &byBarcodeCreated)
	if byBarcodeCreated.Barcode == nil || *byBarcodeCreated.Barcode != "5901234123457" {
		t.Fatalf("expected barcode on created food, got %+v", byBarcodeCreated)
	}

	var barcodeGet struct {
		ID      uint    `json:"id"`
		Barcode *string `json:"barcode"`
	}
	doJSONWithToken(t, http.MethodGet, env.BaseURL+"/api/v1/foods/by-barcode/5901234123457", nil, env.Token, http.StatusOK, &barcodeGet)
	if barcodeGet.ID != byBarcodeCreated.ID {
		t.Fatalf("expected food id %d from barcode lookup, got %d", byBarcodeCreated.ID, barcodeGet.ID)
	}

	doJSONWithToken(t, http.MethodGet, env.BaseURL+"/api/v1/foods/by-barcode/00000000", nil, env.Token, http.StatusNotFound, nil)

	doJSONWithToken(t, http.MethodPost, env.BaseURL+"/api/v1/foods", barcodePayload, env.Token, http.StatusConflict, nil)
}
