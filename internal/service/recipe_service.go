package service

import (
	"context"
	"errors"
	"strings"

	"nutrition/internal/domain/food"
	"nutrition/internal/domain/recipe"
	"nutrition/internal/domain/recipeingredient"
	"nutrition/internal/repository"
)

var (
	ErrRecipeNotFound           = errors.New("recipe not found")
	ErrRecipeForbidden          = errors.New("recipe forbidden")
	ErrInvalidRecipeName        = errors.New("invalid recipe name")
	ErrInvalidYieldWeight       = errors.New("invalid yield weight")
	ErrInvalidRecipeIngredients = errors.New("invalid recipe ingredients")
	ErrIngredientFoodNotFound   = errors.New("ingredient food not found")
)

type RecipeStore interface {
	Create(ctx context.Context, in repository.RecipeCreate) (recipe.Recipe, error)
	GetByID(ctx context.Context, id uint) (recipe.Recipe, error)
	List(ctx context.Context, limit, offset int) ([]recipe.Recipe, error)
	SearchByName(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error)
	Update(ctx context.Context, id uint, in repository.RecipeUpdate) (recipe.Recipe, error)
	Delete(ctx context.Context, id uint) error
}

type FoodReader interface {
	GetByID(ctx context.Context, id uint) (food.Food, error)
}

type RecipeService struct {
	repo       RecipeStore
	foodReader FoodReader
}

type RecipeIngredientInput struct {
	FoodID     uint
	RawWeightG float64
	Position   *int
}

type CreateRecipeInput struct {
	Name         string
	YieldWeightG float64
	Ingredients  []RecipeIngredientInput
}

type UpdateRecipeInput struct {
	Name         *string
	YieldWeightG *float64
	Ingredients  *[]RecipeIngredientInput
}

func NewRecipeService(repo RecipeStore, foodReader FoodReader) *RecipeService {
	return &RecipeService{repo: repo, foodReader: foodReader}
}

func (s *RecipeService) Create(ctx context.Context, userID uint, in CreateRecipeInput) (recipe.Recipe, error) {
	if userID == 0 {
		return recipe.Recipe{}, ErrInvalidUserID
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return recipe.Recipe{}, ErrInvalidRecipeName
	}
	if in.YieldWeightG <= 0 {
		return recipe.Recipe{}, ErrInvalidYieldWeight
	}
	if len(in.Ingredients) == 0 {
		return recipe.Recipe{}, ErrInvalidRecipeIngredients
	}

	kcal, protein, carbs, fat, err := s.calculatePer100g(ctx, in.YieldWeightG, in.Ingredients)
	if err != nil {
		return recipe.Recipe{}, err
	}

	created, err := s.repo.Create(ctx, repository.RecipeCreate{
		UserID:         userID,
		Name:           name,
		YieldWeightG:   in.YieldWeightG,
		KcalPer100g:    kcal,
		ProteinPer100g: protein,
		CarbsPer100g:   carbs,
		FatPer100g:     fat,
		Ingredients:    toRepoIngredients(in.Ingredients),
	})
	if err != nil {
		return recipe.Recipe{}, err
	}

	return created, nil
}

func (s *RecipeService) GetByID(ctx context.Context, id uint) (recipe.Recipe, error) {
	value, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return recipe.Recipe{}, ErrRecipeNotFound
	}
	if err != nil {
		return recipe.Recipe{}, err
	}

	return value, nil
}

func (s *RecipeService) List(ctx context.Context, limit, offset int) ([]recipe.Recipe, error) {
	if !IsValidPagination(limit, offset) {
		return nil, ErrInvalidPagination
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *RecipeService) Search(ctx context.Context, query string, limit, offset int) ([]recipe.Recipe, error) {
	if !IsValidPagination(limit, offset) {
		return nil, ErrInvalidPagination
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return s.repo.List(ctx, limit, offset)
	}
	return s.repo.SearchByName(ctx, q, limit, offset)
}

func (s *RecipeService) Update(ctx context.Context, userID, id uint, in UpdateRecipeInput) (recipe.Recipe, error) {
	if userID == 0 {
		return recipe.Recipe{}, ErrInvalidUserID
	}
	if in.Name == nil && in.YieldWeightG == nil && in.Ingredients == nil {
		return recipe.Recipe{}, ErrNoFieldsToUpdate
	}

	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return recipe.Recipe{}, ErrRecipeNotFound
	}
	if err != nil {
		return recipe.Recipe{}, err
	}
	if existing.UserID != userID {
		return recipe.Recipe{}, ErrRecipeForbidden
	}

	updates := repository.RecipeUpdate{}
	if in.Name != nil {
		trimmed := strings.TrimSpace(*in.Name)
		if trimmed == "" {
			return recipe.Recipe{}, ErrInvalidRecipeName
		}
		updates.Name = &trimmed
	}

	yield := existing.YieldWeightG
	if in.YieldWeightG != nil {
		if *in.YieldWeightG <= 0 {
			return recipe.Recipe{}, ErrInvalidYieldWeight
		}
		yield = *in.YieldWeightG
		updates.YieldWeightG = in.YieldWeightG
	}

	needRecalc := in.YieldWeightG != nil || in.Ingredients != nil
	if in.Ingredients != nil {
		if len(*in.Ingredients) == 0 {
			return recipe.Recipe{}, ErrInvalidRecipeIngredients
		}
		updates.Ingredients = ptrRepoIngredients(*in.Ingredients)
	}

	if needRecalc {
		ingredients := toServiceIngredients(existing.Ingredients)
		if in.Ingredients != nil {
			ingredients = *in.Ingredients
		}

		kcal, protein, carbs, fat, err := s.calculatePer100g(ctx, yield, ingredients)
		if err != nil {
			return recipe.Recipe{}, err
		}
		updates.KcalPer100g = &kcal
		updates.ProteinPer100g = &protein
		updates.CarbsPer100g = &carbs
		updates.FatPer100g = &fat
	}

	value, err := s.repo.Update(ctx, id, updates)
	if errors.Is(err, repository.ErrNotFound) {
		return recipe.Recipe{}, ErrRecipeNotFound
	}
	if err != nil {
		return recipe.Recipe{}, err
	}

	return value, nil
}

func (s *RecipeService) Delete(ctx context.Context, userID, id uint) error {
	if userID == 0 {
		return ErrInvalidUserID
	}

	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrRecipeNotFound
	}
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return ErrRecipeForbidden
	}

	err = s.repo.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrRecipeNotFound
	}
	return err
}

func (s *RecipeService) calculatePer100g(ctx context.Context, yieldWeight float64, ingredients []RecipeIngredientInput) (float64, float64, float64, float64, error) {
	if yieldWeight <= 0 {
		return 0, 0, 0, 0, ErrInvalidYieldWeight
	}
	if len(ingredients) == 0 {
		return 0, 0, 0, 0, ErrInvalidRecipeIngredients
	}

	var totalKcal float64
	var totalProtein float64
	var totalCarbs float64
	var totalFat float64

	for _, item := range ingredients {
		if item.FoodID == 0 || item.RawWeightG <= 0 {
			return 0, 0, 0, 0, ErrInvalidRecipeIngredients
		}

		f, err := s.foodReader.GetByID(ctx, item.FoodID)
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, ErrFoodNotFound) {
			return 0, 0, 0, 0, ErrIngredientFoodNotFound
		}
		if err != nil {
			return 0, 0, 0, 0, err
		}

		ratio := item.RawWeightG / 100.0
		totalKcal += f.KcalPer100g * ratio
		totalProtein += f.ProteinPer100g * ratio
		totalCarbs += f.CarbsPer100g * ratio
		totalFat += f.FatPer100g * ratio
	}

	yieldFactor := yieldWeight / 100.0
	return totalKcal / yieldFactor, totalProtein / yieldFactor, totalCarbs / yieldFactor, totalFat / yieldFactor, nil
}

func toRepoIngredients(items []RecipeIngredientInput) []repository.RecipeIngredientInput {
	out := make([]repository.RecipeIngredientInput, 0, len(items))
	for _, item := range items {
		out = append(out, repository.RecipeIngredientInput{
			FoodID:     item.FoodID,
			RawWeightG: item.RawWeightG,
			Position:   item.Position,
		})
	}
	return out
}

func ptrRepoIngredients(items []RecipeIngredientInput) *[]repository.RecipeIngredientInput {
	mapped := toRepoIngredients(items)
	return &mapped
}

func toServiceIngredients(items []recipeingredient.RecipeIngredient) []RecipeIngredientInput {
	out := make([]RecipeIngredientInput, 0, len(items))
	for _, item := range items {
		out = append(out, RecipeIngredientInput{
			FoodID:     item.FoodID,
			RawWeightG: item.RawWeightG,
			Position:   item.Position,
		})
	}
	return out
}
