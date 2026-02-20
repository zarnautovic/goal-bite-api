//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
)

func TestUsersE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	var out struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodGet, fmt.Sprintf("%s/api/v1/users/%d", env.BaseURL, env.UserID), nil, env.Token, http.StatusOK, &out)
	if out.ID != env.UserID {
		t.Fatalf("expected user id %d, got %d", env.UserID, out.ID)
	}

	updatePayload := map[string]any{
		"name":           "E2E Updated",
		"sex":            "male",
		"birth_date":     "1994-05-18",
		"height_cm":      180.0,
		"activity_level": "active",
	}
	var updated struct {
		ID            uint     `json:"id"`
		Name          string   `json:"name"`
		Sex           *string  `json:"sex"`
		HeightCM      *float64 `json:"height_cm"`
		ActivityLevel *string  `json:"activity_level"`
	}
	doJSONWithToken(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/users/me", env.BaseURL), updatePayload, env.Token, http.StatusOK, &updated)
	if updated.ID != env.UserID || updated.Name != "E2E Updated" {
		t.Fatalf("unexpected updated user: %+v", updated)
	}
	if updated.Sex == nil || *updated.Sex != "male" {
		t.Fatalf("expected sex male, got %+v", updated.Sex)
	}

	clearPayload := map[string]any{
		"activity_level": nil,
	}
	var cleared struct {
		ID            uint    `json:"id"`
		ActivityLevel *string `json:"activity_level"`
	}
	doJSONWithToken(t, http.MethodPatch, fmt.Sprintf("%s/api/v1/users/me", env.BaseURL), clearPayload, env.Token, http.StatusOK, &cleared)
	if cleared.ID != env.UserID {
		t.Fatalf("expected user id %d, got %d", env.UserID, cleared.ID)
	}
	if cleared.ActivityLevel != nil {
		t.Fatalf("expected activity_level to be null after clear, got %+v", cleared.ActivityLevel)
	}
}
