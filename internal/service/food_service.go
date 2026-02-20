package service

import (
	"context"
	"errors"
	"strings"

	"nutrition/internal/domain/food"
	"nutrition/internal/repository"
)

var (
	ErrFoodNotFound         = errors.New("food not found")
	ErrFoodBarcodeNotFound  = errors.New("food barcode not found")
	ErrFoodBarcodeExists    = errors.New("food barcode already exists")
	ErrFoodForbidden        = errors.New("food forbidden")
	ErrInvalidFoodBarcode   = errors.New("invalid food barcode")
	ErrInvalidFoodName      = errors.New("invalid food name")
	ErrInvalidNutritionData = errors.New("invalid nutrition values")
	ErrInvalidPagination    = errors.New("invalid pagination")
	ErrNoFieldsToUpdate     = errors.New("no fields to update")
)

type FoodStore interface {
	Create(ctx context.Context, value food.Food) (food.Food, error)
	GetByID(ctx context.Context, id uint) (food.Food, error)
	GetByBarcode(ctx context.Context, barcode string) (food.Food, error)
	List(ctx context.Context, limit, offset int) ([]food.Food, error)
	SearchByName(ctx context.Context, query string, limit, offset int) ([]food.Food, error)
	Update(ctx context.Context, id uint, updates repository.FoodUpdate) (food.Food, error)
	Delete(ctx context.Context, id uint) error
}

type FoodService struct {
	repo FoodStore
}

type CreateFoodInput struct {
	Name           string
	BrandName      *string
	Barcode        *string
	KcalPer100g    float64
	ProteinPer100g float64
	CarbsPer100g   float64
	FatPer100g     float64
}

type UpdateFoodInput struct {
	Name           *string
	BrandName      *string
	Barcode        *string
	KcalPer100g    *float64
	ProteinPer100g *float64
	CarbsPer100g   *float64
	FatPer100g     *float64
}

func NewFoodService(repo FoodStore) *FoodService {
	return &FoodService{repo: repo}
}

func (s *FoodService) Create(ctx context.Context, userID uint, in CreateFoodInput) (food.Food, error) {
	if userID == 0 {
		return food.Food{}, ErrInvalidUserID
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return food.Food{}, ErrInvalidFoodName
	}
	var brandName *string
	if in.BrandName != nil {
		trimmed := strings.TrimSpace(*in.BrandName)
		if trimmed != "" {
			brandName = &trimmed
		}
	}
	if hasNegative(in.KcalPer100g, in.ProteinPer100g, in.CarbsPer100g, in.FatPer100g) {
		return food.Food{}, ErrInvalidNutritionData
	}
	var barcode *string
	if in.Barcode != nil {
		normalized, ok := normalizeBarcode(*in.Barcode)
		if !ok {
			return food.Food{}, ErrInvalidFoodBarcode
		}
		if _, err := s.repo.GetByBarcode(ctx, normalized); err == nil {
			return food.Food{}, ErrFoodBarcodeExists
		} else if !errors.Is(err, repository.ErrNotFound) {
			return food.Food{}, err
		}
		barcode = &normalized
	}

	value := food.Food{
		UserID:         userID,
		Name:           name,
		BrandName:      brandName,
		Barcode:        barcode,
		KcalPer100g:    in.KcalPer100g,
		ProteinPer100g: in.ProteinPer100g,
		CarbsPer100g:   in.CarbsPer100g,
		FatPer100g:     in.FatPer100g,
	}

	created, err := s.repo.Create(ctx, value)
	if err != nil {
		return food.Food{}, err
	}
	return created, nil
}

func (s *FoodService) GetByID(ctx context.Context, id uint) (food.Food, error) {
	value, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return food.Food{}, ErrFoodNotFound
	}
	if err != nil {
		return food.Food{}, err
	}
	return value, nil
}

func (s *FoodService) GetByBarcode(ctx context.Context, barcodeRaw string) (food.Food, error) {
	barcode, ok := normalizeBarcode(barcodeRaw)
	if !ok {
		return food.Food{}, ErrInvalidFoodBarcode
	}
	value, err := s.repo.GetByBarcode(ctx, barcode)
	if errors.Is(err, repository.ErrNotFound) {
		return food.Food{}, ErrFoodBarcodeNotFound
	}
	if err != nil {
		return food.Food{}, err
	}
	return value, nil
}

func (s *FoodService) List(ctx context.Context, limit, offset int) ([]food.Food, error) {
	if !IsValidPagination(limit, offset) {
		return nil, ErrInvalidPagination
	}
	return s.repo.List(ctx, limit, offset)
}

func (s *FoodService) Search(ctx context.Context, query string, limit, offset int) ([]food.Food, error) {
	if !IsValidPagination(limit, offset) {
		return nil, ErrInvalidPagination
	}
	q := strings.TrimSpace(query)
	if q == "" {
		return s.repo.List(ctx, limit, offset)
	}
	return s.repo.SearchByName(ctx, q, limit, offset)
}

func (s *FoodService) Update(ctx context.Context, userID, id uint, in UpdateFoodInput) (food.Food, error) {
	if userID == 0 {
		return food.Food{}, ErrInvalidUserID
	}
	if in.Name == nil && in.BrandName == nil && in.Barcode == nil && in.KcalPer100g == nil && in.ProteinPer100g == nil && in.CarbsPer100g == nil && in.FatPer100g == nil {
		return food.Food{}, ErrNoFieldsToUpdate
	}
	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return food.Food{}, ErrFoodNotFound
	}
	if err != nil {
		return food.Food{}, err
	}
	if existing.UserID != userID {
		return food.Food{}, ErrFoodForbidden
	}

	updates := repository.FoodUpdate{}
	if in.Name != nil {
		trimmed := strings.TrimSpace(*in.Name)
		if trimmed == "" {
			return food.Food{}, ErrInvalidFoodName
		}
		updates.Name = &trimmed
	}
	if in.BrandName != nil {
		trimmed := strings.TrimSpace(*in.BrandName)
		updates.BrandName = &trimmed
	}
	if in.KcalPer100g != nil {
		if *in.KcalPer100g < 0 {
			return food.Food{}, ErrInvalidNutritionData
		}
		updates.KcalPer100g = in.KcalPer100g
	}
	if in.ProteinPer100g != nil {
		if *in.ProteinPer100g < 0 {
			return food.Food{}, ErrInvalidNutritionData
		}
		updates.ProteinPer100g = in.ProteinPer100g
	}
	if in.CarbsPer100g != nil {
		if *in.CarbsPer100g < 0 {
			return food.Food{}, ErrInvalidNutritionData
		}
		updates.CarbsPer100g = in.CarbsPer100g
	}
	if in.FatPer100g != nil {
		if *in.FatPer100g < 0 {
			return food.Food{}, ErrInvalidNutritionData
		}
		updates.FatPer100g = in.FatPer100g
	}
	if in.Barcode != nil {
		normalized, ok := normalizeBarcode(*in.Barcode)
		if !ok {
			return food.Food{}, ErrInvalidFoodBarcode
		}
		if existingBarcode, err := s.repo.GetByBarcode(ctx, normalized); err == nil {
			if existingBarcode.ID != id {
				return food.Food{}, ErrFoodBarcodeExists
			}
		} else if !errors.Is(err, repository.ErrNotFound) {
			return food.Food{}, err
		}
		updates.Barcode = &normalized
	}

	value, err := s.repo.Update(ctx, id, updates)
	if errors.Is(err, repository.ErrNotFound) {
		return food.Food{}, ErrFoodNotFound
	}
	if err != nil {
		return food.Food{}, err
	}
	return value, nil
}

func (s *FoodService) Delete(ctx context.Context, userID, id uint) error {
	if userID == 0 {
		return ErrInvalidUserID
	}

	existing, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrFoodNotFound
	}
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return ErrFoodForbidden
	}

	err = s.repo.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrFoodNotFound
	}
	return err
}

func hasNegative(values ...float64) bool {
	for _, v := range values {
		if v < 0 {
			return true
		}
	}
	return false
}

func normalizeBarcode(raw string) (string, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	buf := make([]rune, 0, len(trimmed))
	for _, r := range trimmed {
		if r >= '0' && r <= '9' {
			buf = append(buf, r)
			continue
		}
		if r == ' ' || r == '-' {
			continue
		}
		return "", false
	}
	if len(buf) == 0 {
		return "", false
	}
	switch len(buf) {
	case 8, 12, 13, 14:
	default:
		return "", false
	}
	return string(buf), true
}
