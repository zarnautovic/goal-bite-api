//go:build integration

package e2e_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestAuthRegisterLoginE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	email := fmt.Sprintf("new.user+%d@gmail.com", time.Now().UnixNano())
	registerPayload := map[string]any{
		"name":     "New User",
		"email":    email,
		"password": "SuperSecret1!",
	}

	var registerOut struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
		User         struct {
			Email string `json:"email"`
		} `json:"user"`
	}
	doJSON(t, http.MethodPost, env.BaseURL+"/api/v1/auth/register", registerPayload, http.StatusCreated, &registerOut)
	if registerOut.Token == "" {
		t.Fatalf("expected register token")
	}
	if registerOut.RefreshToken == "" {
		t.Fatalf("expected register refresh token")
	}

	loginPayload := map[string]any{
		// dot and plus variant should map to same gmail user.
		"email":    "new.user@gmail.com",
		"password": "SuperSecret1!",
	}
	var loginOut struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	doJSON(t, http.MethodPost, env.BaseURL+"/api/v1/auth/login", loginPayload, http.StatusOK, &loginOut)
	if loginOut.Token == "" {
		t.Fatalf("expected login token")
	}
	if loginOut.RefreshToken == "" {
		t.Fatalf("expected login refresh token")
	}

	var refreshOut struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	doJSON(t, http.MethodPost, env.BaseURL+"/api/v1/auth/refresh", map[string]any{"refresh_token": loginOut.RefreshToken}, http.StatusOK, &refreshOut)
	if refreshOut.Token == "" || refreshOut.RefreshToken == "" {
		t.Fatalf("expected refreshed tokens")
	}

	doJSON(t, http.MethodPost, env.BaseURL+"/api/v1/auth/logout", map[string]any{"refresh_token": refreshOut.RefreshToken}, http.StatusNoContent, nil)

	doJSON(t, http.MethodPost, env.BaseURL+"/api/v1/auth/refresh", map[string]any{"refresh_token": refreshOut.RefreshToken}, http.StatusUnauthorized, nil)

	var healthOut struct {
		Status string `json:"status"`
	}
	doJSONWithToken(t, http.MethodGet, env.BaseURL+"/api/v1/health", nil, refreshOut.Token, http.StatusOK, &healthOut)
	if healthOut.Status != "ok" {
		t.Fatalf("expected status ok, got %q", healthOut.Status)
	}

	var meOut struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
	}
	doJSONWithToken(t, http.MethodGet, env.BaseURL+"/api/v1/auth/me", nil, refreshOut.Token, http.StatusOK, &meOut)
	if meOut.ID == 0 || meOut.Email == "" {
		t.Fatalf("expected current user payload, got %+v", meOut)
	}
}
