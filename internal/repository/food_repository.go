package repository

import (
	"context"
	"errors"
	"strings"

	"goal-bite-api/internal/domain/food"

	"gorm.io/gorm"
)

type FoodRepository struct {
	db *gorm.DB
}

type FoodUpdate struct {
	Name           *string
	BrandName      *string
	Barcode        *string
	KcalPer100g    *float64
	ProteinPer100g *float64
	CarbsPer100g   *float64
	FatPer100g     *float64
}

func NewFoodRepository(database *gorm.DB) *FoodRepository {
	return &FoodRepository{db: database}
}

func (r *FoodRepository) Create(ctx context.Context, value food.Food) (food.Food, error) {
	if err := r.db.WithContext(ctx).Create(&value).Error; err != nil {
		return food.Food{}, err
	}
	return value, nil
}

func (r *FoodRepository) GetByID(ctx context.Context, id uint) (food.Food, error) {
	var f food.Food
	err := r.db.WithContext(ctx).First(&f, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return food.Food{}, ErrNotFound
	}
	if err != nil {
		return food.Food{}, err
	}

	return f, nil
}

func (r *FoodRepository) GetByBarcode(ctx context.Context, barcode string) (food.Food, error) {
	var f food.Food
	err := r.db.WithContext(ctx).Where("barcode = ?", barcode).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return food.Food{}, ErrNotFound
	}
	if err != nil {
		return food.Food{}, err
	}
	return f, nil
}

func (r *FoodRepository) List(ctx context.Context, limit, offset int) ([]food.Food, error) {
	var foods []food.Food
	err := r.db.WithContext(ctx).
		Select("id, user_id, name, brand_name, barcode, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, created_at, updated_at").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&foods).Error
	if err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *FoodRepository) SearchByName(ctx context.Context, query string, limit, offset int) ([]food.Food, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return r.List(ctx, limit, offset)
	}

	var foods []food.Food
	err := r.db.WithContext(ctx).
		Select("id, user_id, name, brand_name, barcode, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, created_at, updated_at").
		Where("name ILIKE ?", "%"+q+"%").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&foods).Error
	if err != nil {
		return nil, err
	}

	return foods, nil
}

func (r *FoodRepository) Update(ctx context.Context, id uint, updates FoodUpdate) (food.Food, error) {
	var f food.Food
	if err := r.db.WithContext(ctx).First(&f, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return food.Food{}, ErrNotFound
		}
		return food.Food{}, err
	}

	changes := map[string]any{}
	if updates.Name != nil {
		changes["name"] = *updates.Name
	}
	if updates.BrandName != nil {
		changes["brand_name"] = *updates.BrandName
	}
	if updates.Barcode != nil {
		changes["barcode"] = *updates.Barcode
	}
	if updates.KcalPer100g != nil {
		changes["kcal_per_100g"] = *updates.KcalPer100g
	}
	if updates.ProteinPer100g != nil {
		changes["protein_per_100g"] = *updates.ProteinPer100g
	}
	if updates.CarbsPer100g != nil {
		changes["carbs_per_100g"] = *updates.CarbsPer100g
	}
	if updates.FatPer100g != nil {
		changes["fat_per_100g"] = *updates.FatPer100g
	}

	if len(changes) > 0 {
		if err := r.db.WithContext(ctx).Model(&f).Updates(changes).Error; err != nil {
			return food.Food{}, err
		}
	}

	if err := r.db.WithContext(ctx).First(&f, id).Error; err != nil {
		return food.Food{}, err
	}

	return f, nil
}

func (r *FoodRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&food.Food{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
