package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nutrition/internal/auth"
	"nutrition/internal/config"
	"nutrition/internal/db"
	httpapi "nutrition/internal/http"
	"nutrition/internal/http/handlers"
	"nutrition/internal/repository"
	"nutrition/internal/service"

	"gorm.io/gorm"
)

type App struct {
	cfg    config.Config
	logger *slog.Logger
	server *http.Server
}

type dbReadinessChecker struct {
	db *gorm.DB
}

func (c dbReadinessChecker) Ready(ctx context.Context) error {
	sqlDB, err := c.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func New(cfg config.Config) (*App, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	database, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	userRepository := repository.NewUserRepository(database)
	userService := service.NewUserService(userRepository)
	jwtManager, err := auth.NewJWTManagerWithKeys(cfg.JWTActiveKID, cfg.JWTKeys)
	if err != nil {
		return nil, fmt.Errorf("init jwt manager: %w", err)
	}
	authSessionRepository := repository.NewAuthSessionRepository(database)
	loginAttempts := service.NewMemoryLoginAttemptTracker(
		cfg.AuthLoginMaxAttempts,
		cfg.AuthLoginAttemptWindow,
		cfg.AuthLoginLockoutWindow,
	)
	authService := service.NewAuthService(userRepository, jwtManager, authSessionRepository, loginAttempts)
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
	readinessChecker := dbReadinessChecker{db: database}
	handler := handlers.New(userService, authService, foodService, recipeService, mealService, bodyWeightLogService, userGoalService, energyService, readinessChecker)
	router := httpapi.NewRouter(handler, logger, jwtManager)
	server := &http.Server{
		Addr:    cfg.Addr(),
		Handler: router,
	}

	return &App{cfg: cfg, logger: logger, server: server}, nil
}

func (a *App) Run() error {
	a.logger.Info("starting api", "addr", a.server.Addr, "env", a.cfg.AppEnv)

	errCh := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-stop:
		a.logger.Info("shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		return fmt.Errorf("http server failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	a.logger.Info("api stopped")
	return nil
}
