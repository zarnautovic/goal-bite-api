package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goal-bite-api/internal/domain/food"
	"goal-bite-api/internal/repository"
	"goal-bite-api/internal/service"
)

type fakeFoodStore struct {
	createFn     func(ctx context.Context, value food.Food) (food.Food, error)
	getFn        func(ctx context.Context, id uint) (food.Food, error)
	getBarcodeFn func(ctx context.Context, barcode string) (food.Food, error)
	listFn       func(ctx context.Context, limit, offset int) ([]food.Food, error)
	searchFn     func(ctx context.Context, query string, limit, offset int) ([]food.Food, error)
	updateFn     func(ctx context.Context, id uint, updates repository.FoodUpdate) (food.Food, error)
	deleteFn     func(ctx context.Context, id uint) error
}

func (f fakeFoodStore) Create(ctx context.Context, value food.Food) (food.Food, error) {
	if f.createFn == nil {
		return value, nil
	}
	return f.createFn(ctx, value)
}

func (f fakeFoodStore) GetByID(ctx context.Context, id uint) (food.Food, error) {
	if f.getFn == nil {
		return food.Food{}, nil
	}
	return f.getFn(ctx, id)
}

func (f fakeFoodStore) GetByBarcode(ctx context.Context, barcode string) (food.Food, error) {
	if f.getBarcodeFn == nil {
		return food.Food{}, repository.ErrNotFound
	}
	return f.getBarcodeFn(ctx, barcode)
}

func (f fakeFoodStore) List(ctx context.Context, limit, offset int) ([]food.Food, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, limit, offset)
}

func (f fakeFoodStore) SearchByName(ctx context.Context, query string, limit, offset int) ([]food.Food, error) {
	if f.searchFn == nil {
		return nil, nil
	}
	return f.searchFn(ctx, query, limit, offset)
}

func (f fakeFoodStore) Update(ctx context.Context, id uint, updates repository.FoodUpdate) (food.Food, error) {
	if f.updateFn == nil {
		return food.Food{}, nil
	}
	return f.updateFn(ctx, id, updates)
}

func (f fakeFoodStore) Delete(ctx context.Context, id uint) error {
	if f.deleteFn == nil {
		return nil
	}
	return f.deleteFn(ctx, id)
}

func TestFoodService(t *testing.T) {
	t.Run("create validates name", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{})
		_, err := svc.Create(context.Background(), 1, service.CreateFoodInput{Name: "  "})
		if !errors.Is(err, service.ErrInvalidFoodName) {
			t.Fatalf("expected ErrInvalidFoodName, got %v", err)
		}
	})

	t.Run("create validates nutrition values", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{})
		_, err := svc.Create(context.Background(), 1, service.CreateFoodInput{Name: "Rice", KcalPer100g: -1})
		if !errors.Is(err, service.ErrInvalidNutritionData) {
			t.Fatalf("expected ErrInvalidNutritionData, got %v", err)
		}
	})

	t.Run("create validates barcode", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{})
		barcode := "ABC"
		_, err := svc.Create(context.Background(), 1, service.CreateFoodInput{Name: "Rice", Barcode: &barcode, KcalPer100g: 130, ProteinPer100g: 2.7, CarbsPer100g: 28, FatPer100g: 0.3})
		if !errors.Is(err, service.ErrInvalidFoodBarcode) {
			t.Fatalf("expected ErrInvalidFoodBarcode, got %v", err)
		}
	})

	t.Run("create trims optional brand name", func(t *testing.T) {
		brand := "  Fage  "
		svc := service.NewFoodService(fakeFoodStore{
			createFn: func(_ context.Context, value food.Food) (food.Food, error) {
				if value.BrandName == nil || *value.BrandName != "Fage" {
					t.Fatalf("expected trimmed brand name, got %#v", value.BrandName)
				}
				return value, nil
			},
		})
		_, err := svc.Create(context.Background(), 1, service.CreateFoodInput{
			Name:           "Rice",
			BrandName:      &brand,
			KcalPer100g:    130,
			ProteinPer100g: 2.7,
			CarbsPer100g:   28,
			FatPer100g:     0.3,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("get by id maps not found", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{}, repository.ErrNotFound
		}})
		_, err := svc.GetByID(context.Background(), 10)
		if !errors.Is(err, service.ErrFoodNotFound) {
			t.Fatalf("expected ErrFoodNotFound, got %v", err)
		}
	})

	t.Run("get by barcode maps not found", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{getBarcodeFn: func(_ context.Context, _ string) (food.Food, error) {
			return food.Food{}, repository.ErrNotFound
		}})
		_, err := svc.GetByBarcode(context.Background(), "5901234123457")
		if !errors.Is(err, service.ErrFoodBarcodeNotFound) {
			t.Fatalf("expected ErrFoodBarcodeNotFound, got %v", err)
		}
	})

	t.Run("list validates pagination", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{})
		_, err := svc.List(context.Background(), 0, 0)
		if !errors.Is(err, service.ErrInvalidPagination) {
			t.Fatalf("expected ErrInvalidPagination, got %v", err)
		}
	})

	t.Run("search validates pagination", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{})
		_, err := svc.Search(context.Background(), "egg", 0, 0)
		if !errors.Is(err, service.ErrInvalidPagination) {
			t.Fatalf("expected ErrInvalidPagination, got %v", err)
		}
	})

	t.Run("search trims query and calls repo search", func(t *testing.T) {
		called := false
		svc := service.NewFoodService(fakeFoodStore{
			searchFn: func(_ context.Context, query string, limit, offset int) ([]food.Food, error) {
				called = true
				if query != "egg" || limit != 20 || offset != 0 {
					t.Fatalf("unexpected args: query=%q limit=%d offset=%d", query, limit, offset)
				}
				return []food.Food{{ID: 1, Name: "Egg"}}, nil
			},
		})
		values, err := svc.Search(context.Background(), "  egg  ", 20, 0)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !called || len(values) != 1 {
			t.Fatalf("expected one value from search")
		}
	})

	t.Run("update requires fields", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{})
		_, err := svc.Update(context.Background(), 1, 1, service.UpdateFoodInput{})
		if !errors.Is(err, service.ErrNoFieldsToUpdate) {
			t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
		}
	})

	t.Run("update success returns updated value", func(t *testing.T) {
		now := time.Now().UTC()
		newName := "New Rice"
		newBrand := "Fage"
		svc := service.NewFoodService(fakeFoodStore{
			getFn: func(_ context.Context, _ uint) (food.Food, error) {
				return food.Food{ID: 1, UserID: 7}, nil
			},
			updateFn: func(_ context.Context, _ uint, updates repository.FoodUpdate) (food.Food, error) {
				return food.Food{ID: 1, Name: *updates.Name, BrandName: updates.BrandName, KcalPer100g: 130, ProteinPer100g: 2.7, CarbsPer100g: 28, FatPer100g: 0.3, CreatedAt: now, UpdatedAt: now}, nil
			},
		})

		got, err := svc.Update(context.Background(), 7, 1, service.UpdateFoodInput{Name: &newName, BrandName: &newBrand})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Name != newName {
			t.Fatalf("expected updated name %q, got %q", newName, got.Name)
		}
		if got.BrandName == nil || *got.BrandName != newBrand {
			t.Fatalf("expected updated brand %q, got %#v", newBrand, got.BrandName)
		}
	})

	t.Run("update forbidden for non owner", func(t *testing.T) {
		name := "New Rice"
		svc := service.NewFoodService(fakeFoodStore{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{ID: 1, UserID: 9}, nil
		}})
		_, err := svc.Update(context.Background(), 7, 1, service.UpdateFoodInput{Name: &name})
		if !errors.Is(err, service.ErrFoodForbidden) {
			t.Fatalf("expected ErrFoodForbidden, got %v", err)
		}
	})

	t.Run("delete forbidden for non owner", func(t *testing.T) {
		svc := service.NewFoodService(fakeFoodStore{getFn: func(_ context.Context, _ uint) (food.Food, error) {
			return food.Food{ID: 1, UserID: 9}, nil
		}})
		err := svc.Delete(context.Background(), 7, 1)
		if !errors.Is(err, service.ErrFoodForbidden) {
			t.Fatalf("expected ErrFoodForbidden, got %v", err)
		}
	})
}
