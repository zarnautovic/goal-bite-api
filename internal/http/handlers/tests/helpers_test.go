package handlers_test

import (
	"context"

	"goal-bite-api/internal/domain/bodyweightlog"
	"goal-bite-api/internal/domain/food"
	"goal-bite-api/internal/domain/meal"
	"goal-bite-api/internal/domain/mealitem"
	"goal-bite-api/internal/domain/recipe"
	"goal-bite-api/internal/domain/usergoal"
	"goal-bite-api/internal/service"
)

type fakeFoodService struct {
	createFn     func(ctx context.Context, userID uint, in service.CreateFoodInput) (food.Food, error)
	getFn        func(ctx context.Context, id uint) (food.Food, error)
	getBarcodeFn func(ctx context.Context, barcode string) (food.Food, error)
	listFn       func(ctx context.Context, limit, offset int) ([]food.Food, error)
	searchFn     func(ctx context.Context, query string, limit, offset int) ([]food.Food, error)
	updateFn     func(ctx context.Context, userID, id uint, in service.UpdateFoodInput) (food.Food, error)
	deleteFn     func(ctx context.Context, userID, id uint) error
}

func (f fakeFoodService) Create(ctx context.Context, userID uint, in service.CreateFoodInput) (food.Food, error) {
	if f.createFn == nil {
		return food.Food{}, nil
	}
	return f.createFn(ctx, userID, in)
}

func (f fakeFoodService) GetByID(ctx context.Context, id uint) (food.Food, error) {
	if f.getFn == nil {
		return food.Food{}, nil
	}
	return f.getFn(ctx, id)
}

func (f fakeFoodService) GetByBarcode(ctx context.Context, barcode string) (food.Food, error) {
	if f.getBarcodeFn == nil {
		return food.Food{}, nil
	}
	return f.getBarcodeFn(ctx, barcode)
}

func (f fakeFoodService) List(ctx context.Context, limit, offset int) ([]food.Food, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, limit, offset)
}

func (f fakeFoodService) Search(ctx context.Context, query string, limit, offset int) ([]food.Food, error) {
	if f.searchFn == nil {
		return nil, nil
	}
	return f.searchFn(ctx, query, limit, offset)
}

func (f fakeFoodService) Update(ctx context.Context, userID, id uint, in service.UpdateFoodInput) (food.Food, error) {
	if f.updateFn == nil {
		return food.Food{}, nil
	}
	return f.updateFn(ctx, userID, id, in)
}

func (f fakeFoodService) Delete(ctx context.Context, userID, id uint) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, userID, id)
}

type fakeRecipeService struct {
	createFn func(ctx context.Context, userID uint, in service.CreateRecipeInput) (recipe.Recipe, error)
	getFn    func(ctx context.Context, id uint) (recipe.Recipe, error)
	listFn   func(ctx context.Context, limit, offset int) ([]recipe.Recipe, error)
	searchFn func(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error)
	updateFn func(ctx context.Context, userID, id uint, in service.UpdateRecipeInput) (recipe.Recipe, error)
	deleteFn func(ctx context.Context, userID, id uint) error
}

func (f fakeRecipeService) Create(ctx context.Context, userID uint, in service.CreateRecipeInput) (recipe.Recipe, error) {
	if f.createFn == nil {
		return recipe.Recipe{}, nil
	}
	return f.createFn(ctx, userID, in)
}

func (f fakeRecipeService) GetByID(ctx context.Context, id uint) (recipe.Recipe, error) {
	if f.getFn == nil {
		return recipe.Recipe{}, nil
	}
	return f.getFn(ctx, id)
}

func (f fakeRecipeService) List(ctx context.Context, limit, offset int) ([]recipe.Recipe, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, limit, offset)
}

func (f fakeRecipeService) Search(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error) {
	if f.searchFn == nil {
		return nil, nil
	}
	return f.searchFn(ctx, query, limit, offset)
}

func (f fakeRecipeService) Update(ctx context.Context, userID, id uint, in service.UpdateRecipeInput) (recipe.Recipe, error) {
	if f.updateFn == nil {
		return recipe.Recipe{}, nil
	}
	return f.updateFn(ctx, userID, id, in)
}

func (f fakeRecipeService) Delete(ctx context.Context, userID, id uint) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, userID, id)
}

type fakeMealService struct {
	createFn     func(ctx context.Context, in service.CreateMealInput) (meal.Meal, error)
	updateFn     func(ctx context.Context, userID, mealID uint, in service.UpdateMealInput) (meal.Meal, error)
	deleteFn     func(ctx context.Context, userID, mealID uint) error
	getFn        func(ctx context.Context, userID, id uint) (meal.Meal, error)
	listFn       func(ctx context.Context, in service.ListMealsInput) ([]meal.Meal, error)
	addItemFn    func(ctx context.Context, userID, mealID uint, in service.AddMealItemInput) (mealitem.MealItem, error)
	updateItemFn func(ctx context.Context, userID, mealID, itemID uint, in service.UpdateMealItemInput) (mealitem.MealItem, error)
	deleteItemFn func(ctx context.Context, userID, mealID, itemID uint) error
	dailyFn      func(ctx context.Context, userID uint, date string) (service.DailyTotalsOutput, error)
}

func (f fakeMealService) Create(ctx context.Context, in service.CreateMealInput) (meal.Meal, error) {
	if f.createFn == nil {
		return meal.Meal{}, nil
	}
	return f.createFn(ctx, in)
}

func (f fakeMealService) Update(ctx context.Context, userID, mealID uint, in service.UpdateMealInput) (meal.Meal, error) {
	if f.updateFn == nil {
		return meal.Meal{}, nil
	}
	return f.updateFn(ctx, userID, mealID, in)
}

func (f fakeMealService) Delete(ctx context.Context, userID, mealID uint) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, userID, mealID)
}

func (f fakeMealService) GetByID(ctx context.Context, userID, id uint) (meal.Meal, error) {
	if f.getFn == nil {
		return meal.Meal{}, nil
	}
	return f.getFn(ctx, userID, id)
}

func (f fakeMealService) List(ctx context.Context, in service.ListMealsInput) ([]meal.Meal, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, in)
}

func (f fakeMealService) AddItem(ctx context.Context, userID, mealID uint, in service.AddMealItemInput) (mealitem.MealItem, error) {
	if f.addItemFn == nil {
		return mealitem.MealItem{}, nil
	}
	return f.addItemFn(ctx, userID, mealID, in)
}

func (f fakeMealService) UpdateItem(ctx context.Context, userID, mealID, itemID uint, in service.UpdateMealItemInput) (mealitem.MealItem, error) {
	if f.updateItemFn == nil {
		return mealitem.MealItem{}, nil
	}
	return f.updateItemFn(ctx, userID, mealID, itemID, in)
}

func (f fakeMealService) DeleteItem(ctx context.Context, userID, mealID, itemID uint) error {
	if f.deleteItemFn == nil {
		return nil
	}
	return f.deleteItemFn(ctx, userID, mealID, itemID)
}

func (f fakeMealService) GetDailyTotals(ctx context.Context, userID uint, date string) (service.DailyTotalsOutput, error) {
	if f.dailyFn == nil {
		return service.DailyTotalsOutput{}, nil
	}
	return f.dailyFn(ctx, userID, date)
}

type fakeBodyWeightLogService struct {
	createFn func(ctx context.Context, in service.CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error)
	listFn   func(ctx context.Context, in service.ListBodyWeightLogsInput) ([]bodyweightlog.BodyWeightLog, error)
	latestFn func(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error)
}

func (f fakeBodyWeightLogService) Create(ctx context.Context, in service.CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error) {
	if f.createFn == nil {
		return bodyweightlog.BodyWeightLog{}, nil
	}
	return f.createFn(ctx, in)
}

func (f fakeBodyWeightLogService) List(ctx context.Context, in service.ListBodyWeightLogsInput) ([]bodyweightlog.BodyWeightLog, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, in)
}

func (f fakeBodyWeightLogService) GetLatest(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error) {
	if f.latestFn == nil {
		return bodyweightlog.BodyWeightLog{}, nil
	}
	return f.latestFn(ctx, userID)
}

type fakeAuthService struct {
	registerFn func(ctx context.Context, in service.RegisterInput) (service.AuthResult, error)
	loginFn    func(ctx context.Context, email, password string) (service.AuthResult, error)
	refreshFn  func(ctx context.Context, refreshToken string) (service.AuthResult, error)
	logoutFn   func(ctx context.Context, refreshToken string) error
}

func (f fakeAuthService) Register(ctx context.Context, in service.RegisterInput) (service.AuthResult, error) {
	if f.registerFn == nil {
		return service.AuthResult{}, nil
	}
	return f.registerFn(ctx, in)
}

func (f fakeAuthService) Login(ctx context.Context, email, password string) (service.AuthResult, error) {
	if f.loginFn == nil {
		return service.AuthResult{}, nil
	}
	return f.loginFn(ctx, email, password)
}

func (f fakeAuthService) Refresh(ctx context.Context, refreshToken string) (service.AuthResult, error) {
	if f.refreshFn == nil {
		return service.AuthResult{}, nil
	}
	return f.refreshFn(ctx, refreshToken)
}

func (f fakeAuthService) Logout(ctx context.Context, refreshToken string) error {
	if f.logoutFn == nil {
		return nil
	}
	return f.logoutFn(ctx, refreshToken)
}

type fakeUserGoalService struct {
	upsertFn   func(ctx context.Context, in service.UpsertUserGoalInput) (usergoal.UserGoal, error)
	getFn      func(ctx context.Context, userID uint) (usergoal.UserGoal, error)
	progressFn func(ctx context.Context, userID uint, date string) (service.DailyProgressOutput, error)
}

func (f fakeUserGoalService) Upsert(ctx context.Context, in service.UpsertUserGoalInput) (usergoal.UserGoal, error) {
	if f.upsertFn == nil {
		return usergoal.UserGoal{}, nil
	}
	return f.upsertFn(ctx, in)
}

func (f fakeUserGoalService) GetByUserID(ctx context.Context, userID uint) (usergoal.UserGoal, error) {
	if f.getFn == nil {
		return usergoal.UserGoal{}, nil
	}
	return f.getFn(ctx, userID)
}

func (f fakeUserGoalService) GetDailyProgress(ctx context.Context, userID uint, date string) (service.DailyProgressOutput, error) {
	if f.progressFn == nil {
		return service.DailyProgressOutput{}, nil
	}
	return f.progressFn(ctx, userID, date)
}

type fakeEnergyService struct {
	progressFn func(ctx context.Context, in service.EnergyProgressInput) (service.EnergyProgressOutput, error)
}

func (f fakeEnergyService) GetProgress(ctx context.Context, in service.EnergyProgressInput) (service.EnergyProgressOutput, error) {
	if f.progressFn == nil {
		return service.EnergyProgressOutput{}, nil
	}
	return f.progressFn(ctx, in)
}
