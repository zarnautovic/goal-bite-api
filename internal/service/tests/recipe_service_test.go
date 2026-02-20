package service_test

import (
	"context"
	"errors"
	"testing"

	"nutrition/internal/domain/food"
	"nutrition/internal/domain/recipe"
	"nutrition/internal/domain/recipeingredient"
	"nutrition/internal/repository"
	"nutrition/internal/service"
)

type fakeRecipeStore struct {
	createFn func(ctx context.Context, in repository.RecipeCreate) (recipe.Recipe, error)
	getFn    func(ctx context.Context, id uint) (recipe.Recipe, error)
	listFn   func(ctx context.Context, limit, offset int) ([]recipe.Recipe, error)
	searchFn func(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error)
	updateFn func(ctx context.Context, id uint, in repository.RecipeUpdate) (recipe.Recipe, error)
	deleteFn func(ctx context.Context, id uint) error
}

func (f fakeRecipeStore) Create(ctx context.Context, in repository.RecipeCreate) (recipe.Recipe, error) {
	if f.createFn == nil {
		return recipe.Recipe{}, nil
	}
	return f.createFn(ctx, in)
}

func (f fakeRecipeStore) GetByID(ctx context.Context, id uint) (recipe.Recipe, error) {
	if f.getFn == nil {
		return recipe.Recipe{}, nil
	}
	return f.getFn(ctx, id)
}

func (f fakeRecipeStore) List(ctx context.Context, limit, offset int) ([]recipe.Recipe, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, limit, offset)
}

func (f fakeRecipeStore) SearchByName(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error) {
	if f.searchFn == nil {
		return nil, nil
	}
	return f.searchFn(ctx, query, limit, offset)
}

func (f fakeRecipeStore) Update(ctx context.Context, id uint, in repository.RecipeUpdate) (recipe.Recipe, error) {
	if f.updateFn == nil {
		return recipe.Recipe{}, nil
	}
	return f.updateFn(ctx, id, in)
}

func (f fakeRecipeStore) Delete(ctx context.Context, id uint) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

type fakeFoodReader struct {
	getFn func(ctx context.Context, id uint) (food.Food, error)
}

func (f fakeFoodReader) GetByID(ctx context.Context, id uint) (food.Food, error) {
	if f.getFn == nil {
		return food.Food{}, nil
	}
	return f.getFn(ctx, id)
}

func TestRecipeServiceCreate(t *testing.T) {
	svc := service.NewRecipeService(
		fakeRecipeStore{createFn: func(_ context.Context, in repository.RecipeCreate) (recipe.Recipe, error) {
			if in.KcalPer100g <= 0 {
				t.Fatalf("expected calculated kcal_per_100g > 0")
			}
			return recipe.Recipe{ID: 1, Name: in.Name, YieldWeightG: in.YieldWeightG, KcalPer100g: in.KcalPer100g, ProteinPer100g: in.ProteinPer100g, CarbsPer100g: in.CarbsPer100g, FatPer100g: in.FatPer100g}, nil
		}},
		fakeFoodReader{getFn: func(_ context.Context, id uint) (food.Food, error) {
			if id == 1 {
				return food.Food{Name: "Beef", KcalPer100g: 250, ProteinPer100g: 26, CarbsPer100g: 0, FatPer100g: 15}, nil
			}
			return food.Food{}, repository.ErrNotFound
		}},
	)

	_, err := svc.Create(context.Background(), 1, service.CreateRecipeInput{
		Name:         "Goulash",
		YieldWeightG: 1000,
		Ingredients: []service.RecipeIngredientInput{
			{FoodID: 1, RawWeightG: 500},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRecipeServiceIngredientFoodMissing(t *testing.T) {
	svc := service.NewRecipeService(
		fakeRecipeStore{},
		fakeFoodReader{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{}, repository.ErrNotFound
		}},
	)

	_, err := svc.Create(context.Background(), 1, service.CreateRecipeInput{
		Name:         "Goulash",
		YieldWeightG: 1000,
		Ingredients: []service.RecipeIngredientInput{
			{FoodID: 999, RawWeightG: 200},
		},
	})
	if !errors.Is(err, service.ErrIngredientFoodNotFound) {
		t.Fatalf("expected ErrIngredientFoodNotFound, got %v", err)
	}
}

func TestRecipeServiceUpdateRecalculates(t *testing.T) {
	svc := service.NewRecipeService(
		fakeRecipeStore{
			getFn: func(_ context.Context, _ uint) (recipe.Recipe, error) {
				return recipe.Recipe{ID: 1, UserID: 7, Name: "Goulash", YieldWeightG: 1000, Ingredients: []recipeingredient.RecipeIngredient{{FoodID: 1, RawWeightG: 500}}}, nil
			},
			updateFn: func(_ context.Context, _ uint, in repository.RecipeUpdate) (recipe.Recipe, error) {
				if in.KcalPer100g == nil || *in.KcalPer100g <= 0 {
					t.Fatalf("expected recalculated kcal_per_100g")
				}
				return recipe.Recipe{ID: 1, Name: "Updated"}, nil
			},
		},
		fakeFoodReader{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{Name: "Beef", KcalPer100g: 250, ProteinPer100g: 26, CarbsPer100g: 0, FatPer100g: 15}, nil
		}},
	)

	name := "Updated"
	yield := 1200.0
	_, err := svc.Update(context.Background(), 7, 1, service.UpdateRecipeInput{Name: &name, YieldWeightG: &yield})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRecipeServiceUpdateForbidden(t *testing.T) {
	svc := service.NewRecipeService(
		fakeRecipeStore{
			getFn: func(_ context.Context, _ uint) (recipe.Recipe, error) {
				return recipe.Recipe{ID: 1, UserID: 9, Name: "Goulash", YieldWeightG: 1000, Ingredients: []recipeingredient.RecipeIngredient{{FoodID: 1, RawWeightG: 500}}}, nil
			},
		},
		fakeFoodReader{},
	)

	name := "Updated"
	_, err := svc.Update(context.Background(), 7, 1, service.UpdateRecipeInput{Name: &name})
	if !errors.Is(err, service.ErrRecipeForbidden) {
		t.Fatalf("expected ErrRecipeForbidden, got %v", err)
	}
}

func TestRecipeServiceDeleteForbidden(t *testing.T) {
	svc := service.NewRecipeService(
		fakeRecipeStore{
			getFn: func(_ context.Context, _ uint) (recipe.Recipe, error) {
				return recipe.Recipe{ID: 1, UserID: 9}, nil
			},
		},
		fakeFoodReader{},
	)

	err := svc.Delete(context.Background(), 7, 1)
	if !errors.Is(err, service.ErrRecipeForbidden) {
		t.Fatalf("expected ErrRecipeForbidden, got %v", err)
	}
}

func TestRecipeServiceSearch(t *testing.T) {
	t.Run("validates pagination", func(t *testing.T) {
		svc := service.NewRecipeService(fakeRecipeStore{}, fakeFoodReader{})
		_, err := svc.Search(context.Background(), "soup", 0, 0)
		if !errors.Is(err, service.ErrInvalidPagination) {
			t.Fatalf("expected ErrInvalidPagination, got %v", err)
		}
	})

	t.Run("trims query and calls repo search", func(t *testing.T) {
		called := false
		svc := service.NewRecipeService(
			fakeRecipeStore{
				searchFn: func(_ context.Context, query string, limit, offset int) ([]recipe.Recipe, error) {
					called = true
					if query != "soup" || limit != 20 || offset != 0 {
						t.Fatalf("unexpected args: query=%q limit=%d offset=%d", query, limit, offset)
					}
					return []recipe.Recipe{{ID: 1, Name: "Soup"}}, nil
				},
			},
			fakeFoodReader{},
		)

		values, err := svc.Search(context.Background(), "  soup  ", 20, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !called || len(values) != 1 {
			t.Fatalf("expected one value from search")
		}
	})
}
