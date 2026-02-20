//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestBodyWeightLogsE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	createBodyWeightLog(t, env.BaseURL, 85.2, env.Token)

	var latestOut struct {
		UserID   uint    `json:"user_id"`
		WeightKG float64 `json:"weight_kg"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/body-weight-logs/latest", env.BaseURL), nil, env.Token, http.StatusOK, &latestOut)
	if latestOut.UserID != env.UserID {
		t.Fatalf("expected user id %d, got %d", env.UserID, latestOut.UserID)
	}
	if latestOut.WeightKG != 85.2 {
		t.Fatalf("expected latest weight 85.2, got %v", latestOut.WeightKG)
	}

	var listOut []struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/body-weight-logs?from=2026-02-01&to=2026-02-28&limit=20&offset=0", env.BaseURL), nil, env.Token, http.StatusOK, &listOut)
	if len(listOut) != 1 {
		t.Fatalf("expected 1 body weight log, got %d", len(listOut))
	}
}
