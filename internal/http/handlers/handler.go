package handlers

import (
	"context"

	"goal-bite-api/internal/domain/bodyweightlog"
	"goal-bite-api/internal/domain/food"
	"goal-bite-api/internal/domain/meal"
	"goal-bite-api/internal/domain/mealitem"
	"goal-bite-api/internal/domain/recipe"
	"goal-bite-api/internal/domain/user"
	"goal-bite-api/internal/domain/usergoal"
	"goal-bite-api/internal/service"
)

type Handler struct {
	userService          UserService
	authService          AuthService
	energyService        EnergyService
	readinessChecker     ReadinessChecker
	foodService          FoodService
	recipeService        RecipeService
	mealService          MealService
	bodyWeightLogService BodyWeightLogService
	userGoalService      UserGoalService
}

type UserService interface {
	GetByID(ctx context.Context, id uint) (user.User, error)
	Update(ctx context.Context, id uint, in service.UpdateUserInput) (user.User, error)
}

type AuthService interface {
	Register(ctx context.Context, in service.RegisterInput) (service.AuthResult, error)
	Login(ctx context.Context, email, password string) (service.AuthResult, error)
	Refresh(ctx context.Context, refreshToken string) (service.AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
}

type FoodService interface {
	Create(ctx context.Context, userID uint, in service.CreateFoodInput) (food.Food, error)
	GetByID(ctx context.Context, id uint) (food.Food, error)
	GetByBarcode(ctx context.Context, barcode string) (food.Food, error)
	List(ctx context.Context, limit, offset int) ([]food.Food, error)
	Search(ctx context.Context, query string, limit, offset int) ([]food.Food, error)
	Update(ctx context.Context, userID, id uint, in service.UpdateFoodInput) (food.Food, error)
	Delete(ctx context.Context, userID, id uint) error
}

type RecipeService interface {
	Create(ctx context.Context, userID uint, in service.CreateRecipeInput) (recipe.Recipe, error)
	GetByID(ctx context.Context, id uint) (recipe.Recipe, error)
	List(ctx context.Context, limit, offset int) ([]recipe.Recipe, error)
	Search(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error)
	Update(ctx context.Context, userID, id uint, in service.UpdateRecipeInput) (recipe.Recipe, error)
	Delete(ctx context.Context, userID, id uint) error
}

type MealService interface {
	Create(ctx context.Context, in service.CreateMealInput) (meal.Meal, error)
	Update(ctx context.Context, userID, mealID uint, in service.UpdateMealInput) (meal.Meal, error)
	Delete(ctx context.Context, userID, mealID uint) error
	GetByID(ctx context.Context, userID, id uint) (meal.Meal, error)
	List(ctx context.Context, in service.ListMealsInput) ([]meal.Meal, error)
	AddItem(ctx context.Context, userID, mealID uint, in service.AddMealItemInput) (mealitem.MealItem, error)
	UpdateItem(ctx context.Context, userID, mealID, itemID uint, in service.UpdateMealItemInput) (mealitem.MealItem, error)
	DeleteItem(ctx context.Context, userID, mealID, itemID uint) error
	GetDailyTotals(ctx context.Context, userID uint, date string) (service.DailyTotalsOutput, error)
}

type BodyWeightLogService interface {
	Create(ctx context.Context, in service.CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error)
	List(ctx context.Context, in service.ListBodyWeightLogsInput) ([]bodyweightlog.BodyWeightLog, error)
	GetLatest(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error)
}

type EnergyService interface {
	GetProgress(ctx context.Context, in service.EnergyProgressInput) (service.EnergyProgressOutput, error)
}

type noopEnergyService struct{}

func (noopEnergyService) GetProgress(_ context.Context, _ service.EnergyProgressInput) (service.EnergyProgressOutput, error) {
	return service.EnergyProgressOutput{}, service.ErrInvalidEnergyProgressQuery
}

type ReadinessChecker interface {
	Ready(ctx context.Context) error
}

type noopReadinessChecker struct{}

func (noopReadinessChecker) Ready(_ context.Context) error {
	return nil
}

type UserGoalService interface {
	Upsert(ctx context.Context, in service.UpsertUserGoalInput) (usergoal.UserGoal, error)
	GetByUserID(ctx context.Context, userID uint) (usergoal.UserGoal, error)
	GetDailyProgress(ctx context.Context, userID uint, date string) (service.DailyProgressOutput, error)
}

type noopUserGoalService struct{}

func (noopUserGoalService) Upsert(_ context.Context, _ service.UpsertUserGoalInput) (usergoal.UserGoal, error) {
	return usergoal.UserGoal{}, service.ErrUserGoalNotFound
}

func (noopUserGoalService) GetByUserID(_ context.Context, _ uint) (usergoal.UserGoal, error) {
	return usergoal.UserGoal{}, service.ErrUserGoalNotFound
}

func (noopUserGoalService) GetDailyProgress(_ context.Context, _ uint, _ string) (service.DailyProgressOutput, error) {
	return service.DailyProgressOutput{}, service.ErrUserGoalNotFound
}

func New(
	userService UserService,
	authService AuthService,
	foodService FoodService,
	recipeService RecipeService,
	mealService MealService,
	bodyWeightLogService BodyWeightLogService,
	opts ...any,
) *Handler {
	energyService := EnergyService(noopEnergyService{})
	userGoalService := UserGoalService(noopUserGoalService{})
	readinessChecker := ReadinessChecker(noopReadinessChecker{})
	for _, opt := range opts {
		switch v := opt.(type) {
		case EnergyService:
			if v != nil {
				energyService = v
			}
		case UserGoalService:
			if v != nil {
				userGoalService = v
			}
		case ReadinessChecker:
			if v != nil {
				readinessChecker = v
			}
		}
	}

	return &Handler{
		userService:          userService,
		authService:          authService,
		energyService:        energyService,
		readinessChecker:     readinessChecker,
		foodService:          foodService,
		recipeService:        recipeService,
		mealService:          mealService,
		bodyWeightLogService: bodyWeightLogService,
		userGoalService:      userGoalService,
	}
}
