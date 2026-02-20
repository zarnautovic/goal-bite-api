//go:build integration

package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nutrition/internal/auth"
	"nutrition/internal/db"
	httpapi "nutrition/internal/http"
	"nutrition/internal/http/handlers"
	"nutrition/internal/repository"
	"nutrition/internal/service"

	"gorm.io/gorm"
)

type testEnv struct {
	BaseURL string
	UserID  uint
	Token   string
	close   func()
}

type testDBReadinessChecker struct {
	db *gorm.DB
}

func (c testDBReadinessChecker) Ready(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

const testJWTSecret = "e2e-test-secret"

func setupTestEnv(t *testing.T) testEnv {
	t.Helper()
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	if testDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	appDatabaseURL := os.Getenv("DATABASE_URL")
	if appDatabaseURL != "" && appDatabaseURL == testDatabaseURL {
		t.Fatalf("unsafe config: TEST_DATABASE_URL must be different from DATABASE_URL")
	}

	database := openTestDB(t, testDatabaseURL)
	migrateUp(t, testDatabaseURL)
	truncateAll(t, database)
	userID := createUser(t, database, "E2E User")
	jwtManager := auth.NewJWTManager(testJWTSecret)
	token, err := jwtManager.Generate(userID)
	if err != nil {
		t.Fatalf("generate jwt: %v", err)
	}

	router := buildRouter(database, jwtManager)
	server := httptest.NewServer(router)

	return testEnv{
		BaseURL: server.URL,
		UserID:  userID,
		Token:   token,
		close:   server.Close,
	}
}

func (e testEnv) Close() {
	if e.close != nil {
		e.close()
	}
}

func openTestDB(t *testing.T, databaseURL string) *gorm.DB {
	t.Helper()
	database, err := db.Open(databaseURL)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	return database
}

func migrateUp(t *testing.T, databaseURL string) {
	t.Helper()
	migrationsPath, err := filepath.Abs("../../internal/db/migrations")
	if err != nil {
		t.Fatalf("resolve migrations path: %v", err)
	}
	if err := db.MigrateUp(databaseURL, "file://"+migrationsPath); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
}

func truncateAll(t *testing.T, database *gorm.DB) {
	t.Helper()
	sql := `
TRUNCATE TABLE
	auth_sessions,
	body_weight_logs,
	meal_items,
	meals,
	recipe_ingredients,
	recipes,
	foods,
	user_goals,
	users
RESTART IDENTITY CASCADE;
`
	if err := database.Exec(sql).Error; err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}

func createUser(t *testing.T, database *gorm.DB, name string) uint {
	t.Helper()
	var id uint
	row := database.Raw(`INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?) RETURNING id`, name, "e2e@example.com", "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy").Row()
	if err := row.Scan(&id); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func buildRouter(database *gorm.DB, jwtManager *auth.JWTManager) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	userRepository := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepository)
	authSessionRepository := repository.NewAuthSessionRepository(database)
	authService := service.NewAuthService(userRepository, jwtManager, authSessionRepository)
	foodRepository := repository.NewFoodRepository(database)
	foodService := service.NewFoodService(foodRepository)
	recipeRepository := repository.NewRecipeRepository(database)
	recipeService := service.NewRecipeService(recipeRepository, foodRepository)
	mealRepository := repository.NewMealRepository(database)
	mealService := service.NewMealService(mealRepository, foodRepository, recipeRepository)
	bodyWeightLogRepository := repository.NewBodyWeightLogRepository(database)
	bodyWeightLogService := service.NewBodyWeightLogService(bodyWeightLogRepository)
	userGoalRepository := repository.NewUserGoalRepository(database)
	userGoalService := service.NewUserGoalService(userGoalRepository, mealRepository)
	energyService := service.NewEnergyService(userRepository, bodyWeightLogRepository, mealRepository)
	handler := handlers.New(
		userService,
		authService,
		foodService,
		recipeService,
		mealService,
		bodyWeightLogService,
		userGoalService,
		energyService,
		testDBReadinessChecker{db: database},
	)
	return httpapi.NewRouter(handler, logger, jwtManager)
}

func createFood(t *testing.T, baseURL, token, name string, kcal, protein, carbs, fat float64) uint {
	t.Helper()
	payload := map[string]any{
		"name":             name,
		"kcal_per_100g":    kcal,
		"protein_per_100g": protein,
		"carbs_per_100g":   carbs,
		"fat_per_100g":     fat,
	}
	var out struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodPost, baseURL+"/api/v1/foods", payload, token, http.StatusCreated, &out)
	return out.ID
}

func createRecipe(t *testing.T, baseURL, token string, foodID uint) uint {
	t.Helper()
	payload := map[string]any{
		"name":           "Rice Bowl",
		"yield_weight_g": 200.0,
		"ingredients": []map[string]any{
			{"food_id": foodID, "raw_weight_g": 200.0},
		},
	}
	var out struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodPost, baseURL+"/api/v1/recipes", payload, token, http.StatusCreated, &out)
	return out.ID
}

func createMealWithFoodItem(t *testing.T, baseURL string, foodID uint, token string) uint {
	t.Helper()
	payload := map[string]any{
		"meal_type": "lunch",
		"eaten_at":  "2026-02-17T12:00:00Z",
		"items": []map[string]any{
			{"food_id": foodID, "weight_g": 150.0},
		},
	}
	var out struct {
		ID uint `json:"id"`
	}
	doJSONWithToken(t, http.MethodPost, baseURL+"/api/v1/meals", payload, token, http.StatusCreated, &out)
	return out.ID
}

func addMealItemRecipe(t *testing.T, baseURL string, mealID, recipeID uint, token string) {
	t.Helper()
	payload := map[string]any{
		"recipe_id": recipeID,
		"weight_g":  120.0,
	}
	doJSONWithToken(t, http.MethodPost, fmt.Sprintf("%s/api/v1/meals/%d/items", baseURL, mealID), payload, token, http.StatusCreated, nil)
}

func createBodyWeightLog(t *testing.T, baseURL string, weightKG float64, token string) {
	t.Helper()
	payload := map[string]any{
		"weight_kg": weightKG,
		"logged_at": "2026-02-17T08:00:00Z",
	}
	doJSONWithToken(t, http.MethodPost, baseURL+"/api/v1/body-weight-logs", payload, token, http.StatusCreated, nil)
}

func upsertUserGoals(t *testing.T, baseURL, token string, targetKcal, targetProteinG, targetCarbsG, targetFatG float64) {
	t.Helper()
	payload := map[string]any{
		"target_kcal":      targetKcal,
		"target_protein_g": targetProteinG,
		"target_carbs_g":   targetCarbsG,
		"target_fat_g":     targetFatG,
	}
	doJSONWithToken(t, http.MethodPut, baseURL+"/api/v1/user-goals", payload, token, http.StatusOK, nil)
}

func doJSON(t *testing.T, method, url string, payload any, expectedStatus int, out any) {
	doJSONWithToken(t, method, url, payload, "", expectedStatus, out)
}

func doJSONWithToken(t *testing.T, method, url string, payload any, token string, expectedStatus int, out any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal payload: %v", err)
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("execute request %s %s: %v", method, url, err)
	}
	defer resp.Body.Close()

	rawResp, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d got %d body=%s", expectedStatus, resp.StatusCode, string(rawResp))
	}

	if out != nil {
		if err := json.Unmarshal(rawResp, out); err != nil {
			t.Fatalf("decode response: %v body=%s", err, string(rawResp))
		}
	}
}
