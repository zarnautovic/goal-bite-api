package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"nutrition/internal/auth"
	"nutrition/internal/http/handlers"
	httpmiddleware "nutrition/internal/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(handler *handlers.Handler, logger *slog.Logger, jwtManager *auth.JWTManager) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(httpmiddleware.NewSlogRequestLogger(logger))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(30 * time.Second))

	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		handlers.WriteErrorResponse(w, handlers.AppError{
			Status:  http.StatusNotFound,
			Code:    "route_not_found",
			Message: "route not found",
		})
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		handlers.WriteErrorResponse(w, handlers.AppError{
			Status:  http.StatusMethodNotAllowed,
			Code:    "method_not_allowed",
			Message: "method not allowed",
		})
	})
	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	router.Route("/api/v1", func(r chi.Router) {
		registerLimiter := httpmiddleware.NewIPRateLimiter(5, time.Minute)
		loginLimiter := httpmiddleware.NewIPRateLimiter(10, time.Minute)
		refreshLimiter := httpmiddleware.NewIPRateLimiter(20, time.Minute)

		r.Get("/health/live", handler.HealthLive)
		r.Get("/health/ready", handler.HealthReady)

		r.With(registerLimiter.Middleware).Post("/auth/register", handler.Register)
		r.With(loginLimiter.Middleware).Post("/auth/login", handler.Login)
		r.With(refreshLimiter.Middleware).Post("/auth/refresh", handler.Refresh)
		r.Post("/auth/logout", handler.Logout)

		r.Group(func(pr chi.Router) {
			pr.Use(httpmiddleware.RequireAuth(jwtManager))
			pr.Get("/auth/me", handler.Me)
			pr.Get("/health", handler.Health)
			pr.Patch("/users/me", handler.UpdateMe)
			pr.Get("/users/{id}", handler.GetUserByID)
			pr.Post("/foods", handler.CreateFood)
			pr.Get("/foods", handler.ListFoods)
			pr.Get("/foods/by-barcode/{barcode}", handler.GetFoodByBarcode)
			pr.Get("/foods/{id}", handler.GetFoodByID)
			pr.Patch("/foods/{id}", handler.UpdateFood)
			pr.Delete("/foods/{id}", handler.DeleteFood)
			pr.Post("/recipes", handler.CreateRecipe)
			pr.Get("/recipes", handler.ListRecipes)
			pr.Get("/recipes/{id}", handler.GetRecipeByID)
			pr.Patch("/recipes/{id}", handler.UpdateRecipe)
			pr.Delete("/recipes/{id}", handler.DeleteRecipe)
			pr.Post("/meals", handler.CreateMeal)
			pr.Get("/meals", handler.ListMeals)
			pr.Get("/meals/{id}", handler.GetMealByID)
			pr.Patch("/meals/{id}", handler.UpdateMeal)
			pr.Delete("/meals/{id}", handler.DeleteMeal)
			pr.Post("/meals/{id}/items", handler.AddMealItem)
			pr.Patch("/meals/{meal_id}/items/{item_id}", handler.UpdateMealItem)
			pr.Delete("/meals/{meal_id}/items/{item_id}", handler.DeleteMealItem)
			pr.Get("/daily-totals", handler.GetDailyTotals)
			pr.Put("/user-goals", handler.UpsertUserGoals)
			pr.Get("/user-goals", handler.GetUserGoals)
			pr.Get("/progress/daily", handler.GetDailyProgress)
			pr.Get("/progress/energy", handler.GetEnergyProgress)
			pr.Post("/body-weight-logs", handler.CreateBodyWeightLog)
			pr.Get("/body-weight-logs", handler.ListBodyWeightLogs)
			pr.Get("/body-weight-logs/latest", handler.GetLatestBodyWeightLog)
		})
	})

	return router
}
