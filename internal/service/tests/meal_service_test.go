package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goal-bite-api/internal/domain/food"
	"goal-bite-api/internal/domain/meal"
	"goal-bite-api/internal/domain/mealitem"
	"goal-bite-api/internal/domain/recipe"
	"goal-bite-api/internal/repository"
	"goal-bite-api/internal/service"
)

type fakeMealStore struct {
	createFn            func(ctx context.Context, in repository.CreateMealInput) (meal.Meal, error)
	createWithItemsFn   func(ctx context.Context, in repository.CreateMealInput, items []repository.AddMealItemInput) (meal.Meal, error)
	getFn               func(ctx context.Context, id uint) (meal.Meal, error)
	getForUserFn        func(ctx context.Context, userID, id uint) (meal.Meal, error)
	listFn              func(ctx context.Context, userID uint, date time.Time, limit, offset int) ([]meal.Meal, error)
	addItemFn           func(ctx context.Context, mealID uint, in repository.AddMealItemInput) (mealitem.MealItem, error)
	addItemForUserFn    func(ctx context.Context, userID, mealID uint, in repository.AddMealItemInput) (mealitem.MealItem, error)
	updateForUserFn     func(ctx context.Context, userID, mealID uint, in repository.UpdateMealInput) (meal.Meal, error)
	deleteForUserFn     func(ctx context.Context, userID, mealID uint) error
	getItemForUserFn    func(ctx context.Context, userID, mealID, itemID uint) (mealitem.MealItem, error)
	updateItemForUserFn func(ctx context.Context, userID, mealID, itemID uint, in repository.AddMealItemInput) (mealitem.MealItem, error)
	deleteItemForUserFn func(ctx context.Context, userID, mealID, itemID uint) error
	dailyFn             func(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error)
}

func (f fakeMealStore) Create(ctx context.Context, in repository.CreateMealInput) (meal.Meal, error) {
	if f.createFn == nil {
		return meal.Meal{}, nil
	}
	return f.createFn(ctx, in)
}

func (f fakeMealStore) CreateWithItems(ctx context.Context, in repository.CreateMealInput, items []repository.AddMealItemInput) (meal.Meal, error) {
	if f.createWithItemsFn == nil {
		return meal.Meal{}, nil
	}
	return f.createWithItemsFn(ctx, in, items)
}

func (f fakeMealStore) GetByID(ctx context.Context, id uint) (meal.Meal, error) {
	if f.getFn == nil {
		return meal.Meal{}, nil
	}
	return f.getFn(ctx, id)
}

func (f fakeMealStore) ListByUserAndDate(ctx context.Context, userID uint, date time.Time, limit, offset int) ([]meal.Meal, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, userID, date, limit, offset)
}

func (f fakeMealStore) GetByIDForUser(ctx context.Context, userID, id uint) (meal.Meal, error) {
	if f.getForUserFn == nil {
		return meal.Meal{}, nil
	}
	return f.getForUserFn(ctx, userID, id)
}

func (f fakeMealStore) AddItem(ctx context.Context, mealID uint, in repository.AddMealItemInput) (mealitem.MealItem, error) {
	if f.addItemFn == nil {
		return mealitem.MealItem{}, nil
	}
	return f.addItemFn(ctx, mealID, in)
}

func (f fakeMealStore) AddItemForUser(ctx context.Context, userID, mealID uint, in repository.AddMealItemInput) (mealitem.MealItem, error) {
	if f.addItemForUserFn == nil {
		return mealitem.MealItem{}, nil
	}
	return f.addItemForUserFn(ctx, userID, mealID, in)
}

func (f fakeMealStore) GetDailyTotals(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error) {
	if f.dailyFn == nil {
		return repository.DailyTotals{}, nil
	}
	return f.dailyFn(ctx, userID, date)
}

func (f fakeMealStore) UpdateForUser(ctx context.Context, userID, mealID uint, in repository.UpdateMealInput) (meal.Meal, error) {
	if f.updateForUserFn == nil {
		return meal.Meal{}, nil
	}
	return f.updateForUserFn(ctx, userID, mealID, in)
}

func (f fakeMealStore) DeleteForUser(ctx context.Context, userID, mealID uint) error {
	if f.deleteForUserFn == nil {
		return nil
	}
	return f.deleteForUserFn(ctx, userID, mealID)
}

func (f fakeMealStore) GetItemForUser(ctx context.Context, userID, mealID, itemID uint) (mealitem.MealItem, error) {
	if f.getItemForUserFn == nil {
		return mealitem.MealItem{}, nil
	}
	return f.getItemForUserFn(ctx, userID, mealID, itemID)
}

func (f fakeMealStore) UpdateItemForUser(ctx context.Context, userID, mealID, itemID uint, in repository.AddMealItemInput) (mealitem.MealItem, error) {
	if f.updateItemForUserFn == nil {
		return mealitem.MealItem{}, nil
	}
	return f.updateItemForUserFn(ctx, userID, mealID, itemID, in)
}

func (f fakeMealStore) DeleteItemForUser(ctx context.Context, userID, mealID, itemID uint) error {
	if f.deleteItemForUserFn == nil {
		return nil
	}
	return f.deleteItemForUserFn(ctx, userID, mealID, itemID)
}

type fakeRecipeReader struct {
	getFn func(ctx context.Context, id uint) (recipe.Recipe, error)
}

func (f fakeRecipeReader) GetByID(ctx context.Context, id uint) (recipe.Recipe, error) {
	if f.getFn == nil {
		return recipe.Recipe{}, nil
	}
	return f.getFn(ctx, id)
}

func TestMealServiceAddItemFoodSnapshot(t *testing.T) {
	fid := uint(1)
	svc := service.NewMealService(
		fakeMealStore{addItemForUserFn: func(_ context.Context, _ uint, _ uint, in repository.AddMealItemInput) (mealitem.MealItem, error) {
			if in.KcalPer100g != 130 {
				t.Fatalf("expected snapshot kcal 130, got %v", in.KcalPer100g)
			}
			return mealitem.MealItem{ID: 1, MealID: 1, FoodID: &fid, WeightG: in.WeightG}, nil
		}},
		fakeFoodStore{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{KcalPer100g: 130, ProteinPer100g: 2.7, CarbsPer100g: 28, FatPer100g: 0.3}, nil
		}},
		fakeRecipeReader{},
	)

	_, err := svc.AddItem(context.Background(), 1, 1, service.AddMealItemInput{FoodID: &fid, WeightG: 200})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMealServiceAddItemXORValidation(t *testing.T) {
	fid := uint(1)
	rid := uint(1)
	svc := service.NewMealService(fakeMealStore{}, fakeFoodStore{}, fakeRecipeReader{})
	_, err := svc.AddItem(context.Background(), 1, 1, service.AddMealItemInput{FoodID: &fid, RecipeID: &rid, WeightG: 100})
	if !errors.Is(err, service.ErrInvalidItemSource) {
		t.Fatalf("expected ErrInvalidItemSource, got %v", err)
	}
}

func TestMealServiceGetDailyTotals(t *testing.T) {
	svc := service.NewMealService(
		fakeMealStore{dailyFn: func(_ context.Context, userID uint, _ time.Time) (repository.DailyTotals, error) {
			if userID != 1 {
				t.Fatalf("expected user id 1, got %d", userID)
			}
			return repository.DailyTotals{Kcal: 1000, Protein: 80, Carbs: 120, Fat: 30}, nil
		}},
		fakeFoodStore{},
		fakeRecipeReader{},
	)

	got, err := svc.GetDailyTotals(context.Background(), 1, "2026-02-17")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.TotalKcal != 1000 || got.TotalProteinG != 80 || got.TotalCarbsG != 120 || got.TotalFatG != 30 {
		t.Fatalf("unexpected totals: %+v", got)
	}
}

func TestMealServiceCreateWithItemsUsesTransactionPath(t *testing.T) {
	fid := uint(1)
	svc := service.NewMealService(
		fakeMealStore{
			createWithItemsFn: func(_ context.Context, _ repository.CreateMealInput, items []repository.AddMealItemInput) (meal.Meal, error) {
				if len(items) != 1 {
					t.Fatalf("expected 1 item, got %d", len(items))
				}
				if items[0].KcalPer100g != 130 {
					t.Fatalf("expected snapshot kcal 130, got %v", items[0].KcalPer100g)
				}
				return meal.Meal{ID: 1, UserID: 1, MealType: meal.MealTypeLunch}, nil
			},
		},
		fakeFoodStore{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{KcalPer100g: 130, ProteinPer100g: 2.7, CarbsPer100g: 28, FatPer100g: 0.3}, nil
		}},
		fakeRecipeReader{},
	)

	_, err := svc.Create(context.Background(), service.CreateMealInput{
		UserID:   1,
		MealType: meal.MealTypeLunch,
		EatenAt:  time.Now().UTC(),
		Items:    []service.AddMealItemInput{{FoodID: &fid, WeightG: 200}},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestMealServiceUpdateItemRecomputesSnapshot(t *testing.T) {
	fid := uint(1)
	itemID := uint(2)
	weight := 180.0
	svc := service.NewMealService(
		fakeMealStore{
			getItemForUserFn: func(_ context.Context, _ uint, _ uint, _ uint) (mealitem.MealItem, error) {
				return mealitem.MealItem{ID: itemID, MealID: 1, FoodID: &fid, WeightG: 100}, nil
			},
			updateItemForUserFn: func(_ context.Context, _ uint, _ uint, _ uint, in repository.AddMealItemInput) (mealitem.MealItem, error) {
				if in.WeightG != 180 {
					t.Fatalf("expected updated weight 180, got %v", in.WeightG)
				}
				if in.KcalPer100g != 130 {
					t.Fatalf("expected snapshot kcal 130, got %v", in.KcalPer100g)
				}
				return mealitem.MealItem{ID: itemID, MealID: 1, FoodID: in.FoodID, WeightG: in.WeightG, KcalPer100g: in.KcalPer100g}, nil
			},
		},
		fakeFoodStore{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{KcalPer100g: 130, ProteinPer100g: 2.7, CarbsPer100g: 28, FatPer100g: 0.3}, nil
		}},
		fakeRecipeReader{},
	)

	_, err := svc.UpdateItem(context.Background(), 1, 1, itemID, service.UpdateMealItemInput{WeightG: &weight})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
