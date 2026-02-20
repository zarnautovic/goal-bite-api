//go:build integration

package e2e_test

import (
	"net/http"
	"testing"
)

func TestHealthE2E(t *testing.T) {
	env := setupTestEnv(t)
	defer env.Close()

	var liveOut struct {
		Status string `json:"status"`
	}
	doJSON(t, http.MethodGet, env.BaseURL+"/api/v1/health/live", nil, http.StatusOK, &liveOut)
	if liveOut.Status != "ok" {
		t.Fatalf("expected status ok, got %q", liveOut.Status)
	}

	var readyOut struct {
		Status string `json:"status"`
	}
	doJSON(t, http.MethodGet, env.BaseURL+"/api/v1/health/ready", nil, http.StatusOK, &readyOut)
	if readyOut.Status != "ok" {
		t.Fatalf("expected status ok, got %q", readyOut.Status)
	}

	var legacyOut struct {
		Status string `json:"status"`
	}
	doJSONWithToken(t, http.MethodGet, env.BaseURL+"/api/v1/health", nil, env.Token, http.StatusOK, &legacyOut)
	if legacyOut.Status != "ok" {
		t.Fatalf("expected status ok, got %q", legacyOut.Status)
	}
}
