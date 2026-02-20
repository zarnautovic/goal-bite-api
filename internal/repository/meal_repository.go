package repository

import (
	"context"
	"errors"
	"time"

	"goal-bite-api/internal/domain/meal"
	"goal-bite-api/internal/domain/mealitem"

	"gorm.io/gorm"
)

type MealRepository struct {
	db *gorm.DB
}

type CreateMealInput struct {
	UserID   uint
	MealType meal.MealType
	EatenAt  time.Time
}

type AddMealItemInput struct {
	FoodID         *uint
	RecipeID       *uint
	WeightG        float64
	KcalPer100g    float64
	ProteinPer100g float64
	CarbsPer100g   float64
	FatPer100g     float64
}

type UpdateMealInput struct {
	MealType *meal.MealType
	EatenAt  *time.Time
}

type DailyTotals struct {
	Kcal    float64
	Protein float64
	Carbs   float64
	Fat     float64
}

func NewMealRepository(database *gorm.DB) *MealRepository {
	return &MealRepository{db: database}
}

func (r *MealRepository) Create(ctx context.Context, in CreateMealInput) (meal.Meal, error) {
	value := meal.Meal{UserID: in.UserID, MealType: in.MealType, EatenAt: in.EatenAt}
	if err := r.db.WithContext(ctx).Create(&value).Error; err != nil {
		return meal.Meal{}, err
	}
	return value, nil
}

func (r *MealRepository) CreateWithItems(ctx context.Context, in CreateMealInput, items []AddMealItemInput) (meal.Meal, error) {
	var out meal.Meal
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		out = meal.Meal{UserID: in.UserID, MealType: in.MealType, EatenAt: in.EatenAt}
		if err := tx.Create(&out).Error; err != nil {
			return err
		}

		dbItems := make([]mealitem.MealItem, 0, len(items))
		for _, inItem := range items {
			dbItems = append(dbItems, mealitem.MealItem{
				MealID:         out.ID,
				FoodID:         inItem.FoodID,
				RecipeID:       inItem.RecipeID,
				WeightG:        inItem.WeightG,
				KcalPer100g:    inItem.KcalPer100g,
				ProteinPer100g: inItem.ProteinPer100g,
				CarbsPer100g:   inItem.CarbsPer100g,
				FatPer100g:     inItem.FatPer100g,
			})
		}

		if err := tx.Create(&dbItems).Error; err != nil {
			return err
		}

		out.Items = dbItems
		return nil
	})
	if err != nil {
		return meal.Meal{}, err
	}
	return out, nil
}

func (r *MealRepository) GetByID(ctx context.Context, id uint) (meal.Meal, error) {
	var out meal.Meal
	if err := r.db.WithContext(ctx).First(&out, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return meal.Meal{}, ErrNotFound
		}
		return meal.Meal{}, err
	}

	var items []mealitem.MealItem
	if err := r.db.WithContext(ctx).
		Where("meal_id = ?", id).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return meal.Meal{}, err
	}
	out.Items = items

	return out, nil
}

func (r *MealRepository) GetByIDForUser(ctx context.Context, userID, id uint) (meal.Meal, error) {
	var out meal.Meal
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).First(&out).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return meal.Meal{}, ErrNotFound
		}
		return meal.Meal{}, err
	}

	var items []mealitem.MealItem
	if err := r.db.WithContext(ctx).
		Where("meal_id = ?", id).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return meal.Meal{}, err
	}
	out.Items = items
	return out, nil
}

func (r *MealRepository) ListByUserAndDate(ctx context.Context, userID uint, date time.Time, limit, offset int) ([]meal.Meal, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	var out []meal.Meal
	err := r.db.WithContext(ctx).
		Select("id, user_id, meal_type, eaten_at, created_at, updated_at").
		Where("user_id = ?", userID).
		Where("eaten_at >= ? AND eaten_at < ?", start, end).
		Order("eaten_at ASC, id ASC").
		Limit(limit).
		Offset(offset).
		Find(&out).Error
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return out, nil
	}

	mealIDs := make([]uint, 0, len(out))
	for _, m := range out {
		mealIDs = append(mealIDs, m.ID)
	}

	var items []mealitem.MealItem
	if err := r.db.WithContext(ctx).
		Select("id, meal_id, food_id, recipe_id, weight_g, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, created_at, updated_at").
		Where("meal_id IN ?", mealIDs).
		Order("id ASC").
		Find(&items).Error; err != nil {
		return nil, err
	}

	itemsByMealID := make(map[uint][]mealitem.MealItem, len(out))
	for _, item := range items {
		itemsByMealID[item.MealID] = append(itemsByMealID[item.MealID], item)
	}
	for i := range out {
		out[i].Items = itemsByMealID[out[i].ID]
	}

	return out, nil
}

func (r *MealRepository) AddItem(ctx context.Context, mealID uint, in AddMealItemInput) (mealitem.MealItem, error) {
	var m meal.Meal
	if err := r.db.WithContext(ctx).First(&m, mealID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mealitem.MealItem{}, ErrNotFound
		}
		return mealitem.MealItem{}, err
	}

	item := mealitem.MealItem{
		MealID:         mealID,
		FoodID:         in.FoodID,
		RecipeID:       in.RecipeID,
		WeightG:        in.WeightG,
		KcalPer100g:    in.KcalPer100g,
		ProteinPer100g: in.ProteinPer100g,
		CarbsPer100g:   in.CarbsPer100g,
		FatPer100g:     in.FatPer100g,
	}
	if err := r.db.WithContext(ctx).Create(&item).Error; err != nil {
		return mealitem.MealItem{}, err
	}

	return item, nil
}

func (r *MealRepository) AddItemForUser(ctx context.Context, userID, mealID uint, in AddMealItemInput) (mealitem.MealItem, error) {
	var m meal.Meal
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", mealID, userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return mealitem.MealItem{}, ErrNotFound
		}
		return mealitem.MealItem{}, err
	}

	item := mealitem.MealItem{
		MealID:         mealID,
		FoodID:         in.FoodID,
		RecipeID:       in.RecipeID,
		WeightG:        in.WeightG,
		KcalPer100g:    in.KcalPer100g,
		ProteinPer100g: in.ProteinPer100g,
		CarbsPer100g:   in.CarbsPer100g,
		FatPer100g:     in.FatPer100g,
	}
	if err := r.db.WithContext(ctx).Create(&item).Error; err != nil {
		return mealitem.MealItem{}, err
	}

	return item, nil
}

func (r *MealRepository) UpdateForUser(ctx context.Context, userID, mealID uint, in UpdateMealInput) (meal.Meal, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing meal.Meal
		if err := tx.Where("id = ? AND user_id = ?", mealID, userID).First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		updates := map[string]any{}
		if in.MealType != nil {
			updates["meal_type"] = *in.MealType
		}
		if in.EatenAt != nil {
			updates["eaten_at"] = *in.EatenAt
		}
		if len(updates) == 0 {
			return nil
		}

		if err := tx.Model(&meal.Meal{}).Where("id = ? AND user_id = ?", mealID, userID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return meal.Meal{}, err
	}

	return r.GetByIDForUser(ctx, userID, mealID)
}

func (r *MealRepository) DeleteForUser(ctx context.Context, userID, mealID uint) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", mealID, userID).Delete(&meal.Meal{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *MealRepository) GetItemForUser(ctx context.Context, userID, mealID, itemID uint) (mealitem.MealItem, error) {
	var out mealitem.MealItem
	err := r.db.WithContext(ctx).
		Model(&mealitem.MealItem{}).
		Joins("JOIN meals m ON m.id = meal_items.meal_id").
		Where("meal_items.id = ? AND meal_items.meal_id = ? AND m.user_id = ?", itemID, mealID, userID).
		First(&out).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mealitem.MealItem{}, ErrNotFound
	}
	if err != nil {
		return mealitem.MealItem{}, err
	}
	return out, nil
}

func (r *MealRepository) UpdateItemForUser(ctx context.Context, userID, mealID, itemID uint, in AddMealItemInput) (mealitem.MealItem, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing mealitem.MealItem
		if err := tx.Model(&mealitem.MealItem{}).
			Joins("JOIN meals m ON m.id = meal_items.meal_id").
			Where("meal_items.id = ? AND meal_items.meal_id = ? AND m.user_id = ?", itemID, mealID, userID).
			First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		updates := map[string]any{
			"food_id":          in.FoodID,
			"recipe_id":        in.RecipeID,
			"weight_g":         in.WeightG,
			"kcal_per_100g":    in.KcalPer100g,
			"protein_per_100g": in.ProteinPer100g,
			"carbs_per_100g":   in.CarbsPer100g,
			"fat_per_100g":     in.FatPer100g,
		}
		if err := tx.Model(&mealitem.MealItem{}).Where("id = ?", itemID).Updates(updates).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return mealitem.MealItem{}, err
	}

	return r.GetItemForUser(ctx, userID, mealID, itemID)
}

func (r *MealRepository) DeleteItemForUser(ctx context.Context, userID, mealID, itemID uint) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing mealitem.MealItem
		if err := tx.Model(&mealitem.MealItem{}).
			Joins("JOIN meals m ON m.id = meal_items.meal_id").
			Where("meal_items.id = ? AND meal_items.meal_id = ? AND m.user_id = ?", itemID, mealID, userID).
			First(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}
		if err := tx.Where("id = ?", itemID).Delete(&mealitem.MealItem{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MealRepository) GetDailyTotals(ctx context.Context, userID uint, date time.Time) (DailyTotals, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	var out DailyTotals
	row := r.db.WithContext(ctx).Raw(`
		SELECT
			COALESCE(SUM(mi.kcal_per_100g * mi.weight_g / 100.0), 0) AS kcal,
			COALESCE(SUM(mi.protein_per_100g * mi.weight_g / 100.0), 0) AS protein,
			COALESCE(SUM(mi.carbs_per_100g * mi.weight_g / 100.0), 0) AS carbs,
			COALESCE(SUM(mi.fat_per_100g * mi.weight_g / 100.0), 0) AS fat
		FROM meal_items mi
		JOIN meals m ON m.id = mi.meal_id
		WHERE m.user_id = ? AND m.eaten_at >= ? AND m.eaten_at < ?
	`, userID, start, end).Row()

	if err := row.Scan(&out.Kcal, &out.Protein, &out.Carbs, &out.Fat); err != nil {
		return DailyTotals{}, err
	}
	return out, nil
}
