package service

import (
	"context"
	"errors"
	"time"

	"goal-bite-api/internal/domain/meal"
	"goal-bite-api/internal/domain/mealitem"
	"goal-bite-api/internal/domain/recipe"
	"goal-bite-api/internal/repository"
)

var (
	ErrMealNotFound         = errors.New("meal not found")
	ErrMealItemNotFound     = errors.New("meal item not found")
	ErrInvalidMealType      = errors.New("invalid meal type")
	ErrInvalidEatenAt       = errors.New("invalid eaten_at")
	ErrInvalidUserID        = errors.New("invalid user id")
	ErrInvalidItemSource    = errors.New("invalid item source")
	ErrInvalidItemWeight    = errors.New("invalid item weight")
	ErrInvalidDate          = errors.New("invalid date")
	ErrRecipeSourceNotFound = errors.New("recipe source not found")
)

type MealStore interface {
	Create(ctx context.Context, in repository.CreateMealInput) (meal.Meal, error)
	CreateWithItems(ctx context.Context, in repository.CreateMealInput, items []repository.AddMealItemInput) (meal.Meal, error)
	GetByID(ctx context.Context, id uint) (meal.Meal, error)
	GetByIDForUser(ctx context.Context, userID, id uint) (meal.Meal, error)
	ListByUserAndDate(ctx context.Context, userID uint, date time.Time, limit, offset int) ([]meal.Meal, error)
	AddItem(ctx context.Context, mealID uint, in repository.AddMealItemInput) (mealitem.MealItem, error)
	AddItemForUser(ctx context.Context, userID, mealID uint, in repository.AddMealItemInput) (mealitem.MealItem, error)
	UpdateForUser(ctx context.Context, userID, mealID uint, in repository.UpdateMealInput) (meal.Meal, error)
	DeleteForUser(ctx context.Context, userID, mealID uint) error
	GetItemForUser(ctx context.Context, userID, mealID, itemID uint) (mealitem.MealItem, error)
	UpdateItemForUser(ctx context.Context, userID, mealID, itemID uint, in repository.AddMealItemInput) (mealitem.MealItem, error)
	DeleteItemForUser(ctx context.Context, userID, mealID, itemID uint) error
	GetDailyTotals(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error)
}

type RecipeReader interface {
	GetByID(ctx context.Context, id uint) (recipe.Recipe, error)
}

type MealService struct {
	repo         MealStore
	foodReader   FoodReader
	recipeReader RecipeReader
}

type CreateMealInput struct {
	UserID   uint
	MealType meal.MealType
	EatenAt  time.Time
	Items    []AddMealItemInput
}

type AddMealItemInput struct {
	FoodID   *uint
	RecipeID *uint
	WeightG  float64
}

type ListMealsInput struct {
	UserID uint
	Date   string
	Limit  int
	Offset int
}

type UpdateMealInput struct {
	MealType *meal.MealType
	EatenAt  *time.Time
}

type UpdateMealItemInput struct {
	FoodID   *uint
	RecipeID *uint
	WeightG  *float64
}

type DailyTotalsOutput struct {
	Date          string  `json:"date"`
	TotalKcal     float64 `json:"total_kcal"`
	TotalProteinG float64 `json:"total_protein_g"`
	TotalCarbsG   float64 `json:"total_carbs_g"`
	TotalFatG     float64 `json:"total_fat_g"`
}

func NewMealService(repo MealStore, foodReader FoodReader, recipeReader RecipeReader) *MealService {
	return &MealService{repo: repo, foodReader: foodReader, recipeReader: recipeReader}
}

func (s *MealService) Create(ctx context.Context, in CreateMealInput) (meal.Meal, error) {
	if in.UserID == 0 {
		return meal.Meal{}, ErrInvalidUserID
	}
	if !isValidMealType(in.MealType) {
		return meal.Meal{}, ErrInvalidMealType
	}
	if in.EatenAt.IsZero() {
		return meal.Meal{}, ErrInvalidEatenAt
	}

	if len(in.Items) > 0 {
		snapshots := make([]repository.AddMealItemInput, 0, len(in.Items))
		for _, item := range in.Items {
			snapshot, err := s.resolveMealItemSnapshot(ctx, item)
			if err != nil {
				return meal.Meal{}, err
			}
			snapshots = append(snapshots, snapshot)
		}

		value, err := s.repo.CreateWithItems(ctx, repository.CreateMealInput{
			UserID:   in.UserID,
			MealType: in.MealType,
			EatenAt:  in.EatenAt.UTC(),
		}, snapshots)
		if err != nil {
			return meal.Meal{}, err
		}
		return value, nil
	}

	value, err := s.repo.Create(ctx, repository.CreateMealInput{UserID: in.UserID, MealType: in.MealType, EatenAt: in.EatenAt.UTC()})
	if err != nil {
		return meal.Meal{}, err
	}
	return value, nil
}

func (s *MealService) GetByID(ctx context.Context, userID, id uint) (meal.Meal, error) {
	if userID == 0 {
		return meal.Meal{}, ErrInvalidUserID
	}
	value, err := s.repo.GetByIDForUser(ctx, userID, id)
	if errors.Is(err, repository.ErrNotFound) {
		return meal.Meal{}, ErrMealNotFound
	}
	if err != nil {
		return meal.Meal{}, err
	}
	return value, nil
}

func (s *MealService) List(ctx context.Context, in ListMealsInput) ([]meal.Meal, error) {
	if in.UserID == 0 {
		return nil, ErrInvalidUserID
	}
	if !IsValidPagination(in.Limit, in.Offset) {
		return nil, ErrInvalidPagination
	}
	date, err := time.Parse("2006-01-02", in.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}
	return s.repo.ListByUserAndDate(ctx, in.UserID, date.UTC(), in.Limit, in.Offset)
}

func (s *MealService) AddItem(ctx context.Context, userID, mealID uint, in AddMealItemInput) (mealitem.MealItem, error) {
	if userID == 0 {
		return mealitem.MealItem{}, ErrInvalidUserID
	}
	if mealID == 0 {
		return mealitem.MealItem{}, ErrMealNotFound
	}
	snapshot, err := s.resolveMealItemSnapshot(ctx, in)
	if err != nil {
		return mealitem.MealItem{}, err
	}

	value, err := s.repo.AddItemForUser(ctx, userID, mealID, snapshot)
	if errors.Is(err, repository.ErrNotFound) {
		return mealitem.MealItem{}, ErrMealNotFound
	}
	if err != nil {
		return mealitem.MealItem{}, err
	}

	return value, nil
}

func (s *MealService) Update(ctx context.Context, userID, mealID uint, in UpdateMealInput) (meal.Meal, error) {
	if userID == 0 {
		return meal.Meal{}, ErrInvalidUserID
	}
	if mealID == 0 {
		return meal.Meal{}, ErrMealNotFound
	}
	if in.MealType == nil && in.EatenAt == nil {
		return meal.Meal{}, ErrNoFieldsToUpdate
	}
	if in.MealType != nil && !isValidMealType(*in.MealType) {
		return meal.Meal{}, ErrInvalidMealType
	}
	if in.EatenAt != nil && in.EatenAt.IsZero() {
		return meal.Meal{}, ErrInvalidEatenAt
	}

	repoIn := repository.UpdateMealInput{
		MealType: in.MealType,
	}
	if in.EatenAt != nil {
		v := in.EatenAt.UTC()
		repoIn.EatenAt = &v
	}

	value, err := s.repo.UpdateForUser(ctx, userID, mealID, repoIn)
	if errors.Is(err, repository.ErrNotFound) {
		return meal.Meal{}, ErrMealNotFound
	}
	if err != nil {
		return meal.Meal{}, err
	}
	return value, nil
}

func (s *MealService) Delete(ctx context.Context, userID, mealID uint) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	if mealID == 0 {
		return ErrMealNotFound
	}
	err := s.repo.DeleteForUser(ctx, userID, mealID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrMealNotFound
	}
	return err
}

func (s *MealService) UpdateItem(ctx context.Context, userID, mealID, itemID uint, in UpdateMealItemInput) (mealitem.MealItem, error) {
	if userID == 0 {
		return mealitem.MealItem{}, ErrInvalidUserID
	}
	if mealID == 0 {
		return mealitem.MealItem{}, ErrMealNotFound
	}
	if itemID == 0 {
		return mealitem.MealItem{}, ErrMealItemNotFound
	}
	if in.FoodID == nil && in.RecipeID == nil && in.WeightG == nil {
		return mealitem.MealItem{}, ErrNoFieldsToUpdate
	}
	if in.WeightG != nil && *in.WeightG <= 0 {
		return mealitem.MealItem{}, ErrInvalidItemWeight
	}

	sourceProvided := in.FoodID != nil || in.RecipeID != nil
	if sourceProvided && (in.FoodID != nil && in.RecipeID != nil) {
		return mealitem.MealItem{}, ErrInvalidItemSource
	}

	existing, err := s.repo.GetItemForUser(ctx, userID, mealID, itemID)
	if errors.Is(err, repository.ErrNotFound) {
		return mealitem.MealItem{}, ErrMealItemNotFound
	}
	if err != nil {
		return mealitem.MealItem{}, err
	}

	finalWeight := existing.WeightG
	if in.WeightG != nil {
		finalWeight = *in.WeightG
	}

	finalFoodID := existing.FoodID
	finalRecipeID := existing.RecipeID
	if sourceProvided {
		finalFoodID = in.FoodID
		finalRecipeID = in.RecipeID
	}

	snapshot, err := s.resolveMealItemSnapshot(ctx, AddMealItemInput{
		FoodID:   finalFoodID,
		RecipeID: finalRecipeID,
		WeightG:  finalWeight,
	})
	if err != nil {
		return mealitem.MealItem{}, err
	}

	value, err := s.repo.UpdateItemForUser(ctx, userID, mealID, itemID, snapshot)
	if errors.Is(err, repository.ErrNotFound) {
		return mealitem.MealItem{}, ErrMealItemNotFound
	}
	if err != nil {
		return mealitem.MealItem{}, err
	}
	return value, nil
}

func (s *MealService) DeleteItem(ctx context.Context, userID, mealID, itemID uint) error {
	if userID == 0 {
		return ErrInvalidUserID
	}
	if mealID == 0 {
		return ErrMealNotFound
	}
	if itemID == 0 {
		return ErrMealItemNotFound
	}
	err := s.repo.DeleteItemForUser(ctx, userID, mealID, itemID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrMealItemNotFound
	}
	return err
}

func (s *MealService) resolveMealItemSnapshot(ctx context.Context, in AddMealItemInput) (repository.AddMealItemInput, error) {
	if in.WeightG <= 0 {
		return repository.AddMealItemInput{}, ErrInvalidItemWeight
	}

	foodSet := in.FoodID != nil
	recipeSet := in.RecipeID != nil
	if foodSet == recipeSet {
		return repository.AddMealItemInput{}, ErrInvalidItemSource
	}

	var kcal, protein, carbs, fat float64
	if foodSet {
		f, err := s.foodReader.GetByID(ctx, *in.FoodID)
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, ErrFoodNotFound) {
			return repository.AddMealItemInput{}, ErrFoodNotFound
		}
		if err != nil {
			return repository.AddMealItemInput{}, err
		}
		kcal, protein, carbs, fat = f.KcalPer100g, f.ProteinPer100g, f.CarbsPer100g, f.FatPer100g
	}
	if recipeSet {
		rv, err := s.recipeReader.GetByID(ctx, *in.RecipeID)
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, ErrRecipeNotFound) {
			return repository.AddMealItemInput{}, ErrRecipeSourceNotFound
		}
		if err != nil {
			return repository.AddMealItemInput{}, err
		}
		kcal, protein, carbs, fat = rv.KcalPer100g, rv.ProteinPer100g, rv.CarbsPer100g, rv.FatPer100g
	}

	return repository.AddMealItemInput{
		FoodID:         in.FoodID,
		RecipeID:       in.RecipeID,
		WeightG:        in.WeightG,
		KcalPer100g:    kcal,
		ProteinPer100g: protein,
		CarbsPer100g:   carbs,
		FatPer100g:     fat,
	}, nil
}

func (s *MealService) GetDailyTotals(ctx context.Context, userID uint, date string) (DailyTotalsOutput, error) {
	if userID == 0 {
		return DailyTotalsOutput{}, ErrInvalidUserID
	}
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return DailyTotalsOutput{}, ErrInvalidDate
	}

	totals, err := s.repo.GetDailyTotals(ctx, userID, parsedDate.UTC())
	if err != nil {
		return DailyTotalsOutput{}, err
	}

	return DailyTotalsOutput{
		Date:          parsedDate.Format("2006-01-02"),
		TotalKcal:     totals.Kcal,
		TotalProteinG: totals.Protein,
		TotalCarbsG:   totals.Carbs,
		TotalFatG:     totals.Fat,
	}, nil
}

func isValidMealType(value meal.MealType) bool {
	switch value {
	case meal.MealTypeBreakfast, meal.MealTypeLunch, meal.MealTypeDinner, meal.MealTypeSnack:
		return true
	default:
		return false
	}
}
