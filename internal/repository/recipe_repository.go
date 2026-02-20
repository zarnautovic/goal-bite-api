package repository

import (
	"context"
	"errors"
	"strings"

	"goal-bite-api/internal/domain/recipe"
	"goal-bite-api/internal/domain/recipeingredient"

	"gorm.io/gorm"
)

type RecipeRepository struct {
	db *gorm.DB
}

type RecipeIngredientInput struct {
	FoodID     uint
	RawWeightG float64
	Position   *int
}

type RecipeCreate struct {
	UserID         uint
	Name           string
	YieldWeightG   float64
	KcalPer100g    float64
	ProteinPer100g float64
	CarbsPer100g   float64
	FatPer100g     float64
	Ingredients    []RecipeIngredientInput
}

type RecipeUpdate struct {
	Name           *string
	YieldWeightG   *float64
	KcalPer100g    *float64
	ProteinPer100g *float64
	CarbsPer100g   *float64
	FatPer100g     *float64
	Ingredients    *[]RecipeIngredientInput
}

func NewRecipeRepository(database *gorm.DB) *RecipeRepository {
	return &RecipeRepository{db: database}
}

func (r *RecipeRepository) Create(ctx context.Context, in RecipeCreate) (recipe.Recipe, error) {
	var out recipe.Recipe
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		value := recipe.Recipe{
			UserID:         in.UserID,
			Name:           in.Name,
			YieldWeightG:   in.YieldWeightG,
			KcalPer100g:    in.KcalPer100g,
			ProteinPer100g: in.ProteinPer100g,
			CarbsPer100g:   in.CarbsPer100g,
			FatPer100g:     in.FatPer100g,
		}
		if err := tx.Create(&value).Error; err != nil {
			return err
		}

		ingredients := make([]recipeingredient.RecipeIngredient, 0, len(in.Ingredients))
		for _, item := range in.Ingredients {
			ingredients = append(ingredients, recipeingredient.RecipeIngredient{
				RecipeID:   value.ID,
				FoodID:     item.FoodID,
				RawWeightG: item.RawWeightG,
				Position:   item.Position,
			})
		}
		if len(ingredients) > 0 {
			if err := tx.Create(&ingredients).Error; err != nil {
				return err
			}
		}

		value.Ingredients = ingredients
		out = value
		return nil
	})
	if err != nil {
		return recipe.Recipe{}, err
	}

	return out, nil
}

func (r *RecipeRepository) GetByID(ctx context.Context, id uint) (recipe.Recipe, error) {
	var out recipe.Recipe
	if err := r.db.WithContext(ctx).First(&out, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return recipe.Recipe{}, ErrNotFound
		}
		return recipe.Recipe{}, err
	}

	var ingredients []recipeingredient.RecipeIngredient
	if err := r.db.WithContext(ctx).
		Where("recipe_id = ?", id).
		Order("position ASC NULLS LAST, id ASC").
		Find(&ingredients).Error; err != nil {
		return recipe.Recipe{}, err
	}
	out.Ingredients = ingredients

	return out, nil
}

func (r *RecipeRepository) List(ctx context.Context, limit, offset int) ([]recipe.Recipe, error) {
	var out []recipe.Recipe
	err := r.db.WithContext(ctx).
		Select("id, user_id, name, yield_weight_g, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, created_at, updated_at").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *RecipeRepository) SearchByName(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return r.List(ctx, limit, offset)
	}

	var out []recipe.Recipe
	err := r.db.WithContext(ctx).
		Select("id, user_id, name, yield_weight_g, kcal_per_100g, protein_per_100g, carbs_per_100g, fat_per_100g, created_at, updated_at").
		Where("name ILIKE ?", "%"+q+"%").
		Order("id ASC").
		Limit(limit).
		Offset(offset).
		Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *RecipeRepository) Update(ctx context.Context, id uint, in RecipeUpdate) (recipe.Recipe, error) {
	var out recipe.Recipe
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&out, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return err
		}

		changes := map[string]any{}
		if in.Name != nil {
			changes["name"] = *in.Name
		}
		if in.YieldWeightG != nil {
			changes["yield_weight_g"] = *in.YieldWeightG
		}
		if in.KcalPer100g != nil {
			changes["kcal_per_100g"] = *in.KcalPer100g
		}
		if in.ProteinPer100g != nil {
			changes["protein_per_100g"] = *in.ProteinPer100g
		}
		if in.CarbsPer100g != nil {
			changes["carbs_per_100g"] = *in.CarbsPer100g
		}
		if in.FatPer100g != nil {
			changes["fat_per_100g"] = *in.FatPer100g
		}

		if len(changes) > 0 {
			if err := tx.Model(&out).Updates(changes).Error; err != nil {
				return err
			}
		}

		if in.Ingredients != nil {
			if err := tx.Where("recipe_id = ?", id).Delete(&recipeingredient.RecipeIngredient{}).Error; err != nil {
				return err
			}

			ingredients := make([]recipeingredient.RecipeIngredient, 0, len(*in.Ingredients))
			for _, item := range *in.Ingredients {
				ingredients = append(ingredients, recipeingredient.RecipeIngredient{
					RecipeID:   id,
					FoodID:     item.FoodID,
					RawWeightG: item.RawWeightG,
					Position:   item.Position,
				})
			}
			if len(ingredients) > 0 {
				if err := tx.Create(&ingredients).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return recipe.Recipe{}, ErrNotFound
		}
		return recipe.Recipe{}, err
	}

	return r.GetByID(ctx, id)
}

func (r *RecipeRepository) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&recipe.Recipe{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
