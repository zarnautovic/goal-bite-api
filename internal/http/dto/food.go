package dto

import (
	"errors"
	"strings"

	"nutrition/internal/service"
)

var (
	ErrInvalidName      = errors.New("invalid name")
	ErrInvalidNutrition = errors.New("invalid nutrition values")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)

type CreateFoodRequest struct {
	// Human-readable food name.
	Name string `json:"name" example:"Rice"`
	// Optional brand name.
	BrandName *string `json:"brand_name,omitempty" example:"Fage"`
	// Optional product barcode (EAN/UPC digits).
	Barcode *string `json:"barcode,omitempty" example:"5901234123457"`
	// Energy in kcal per 100g.
	KcalPer100g float64 `json:"kcal_per_100g" example:"130"`
	// Protein grams per 100g.
	ProteinPer100g float64 `json:"protein_per_100g" example:"2.7"`
	// Carbohydrate grams per 100g.
	CarbsPer100g float64 `json:"carbs_per_100g" example:"28"`
	// Fat grams per 100g.
	FatPer100g float64 `json:"fat_per_100g" example:"0.3"`
}

func (r *CreateFoodRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrInvalidName
	}
	if r.KcalPer100g < 0 || r.ProteinPer100g < 0 || r.CarbsPer100g < 0 || r.FatPer100g < 0 {
		return ErrInvalidNutrition
	}
	return nil
}

func (r *CreateFoodRequest) ToServiceInput() service.CreateFoodInput {
	return service.CreateFoodInput{
		Name:           r.Name,
		BrandName:      r.BrandName,
		Barcode:        r.Barcode,
		KcalPer100g:    r.KcalPer100g,
		ProteinPer100g: r.ProteinPer100g,
		CarbsPer100g:   r.CarbsPer100g,
		FatPer100g:     r.FatPer100g,
	}
}

type UpdateFoodRequest struct {
	// Optional food name.
	Name *string `json:"name" example:"Cooked Rice"`
	// Optional brand name.
	BrandName *string `json:"brand_name,omitempty" example:"Fage"`
	// Optional product barcode (EAN/UPC digits).
	Barcode *string `json:"barcode,omitempty" example:"5901234123457"`
	// Optional energy in kcal per 100g.
	KcalPer100g *float64 `json:"kcal_per_100g" example:"130"`
	// Optional protein grams per 100g.
	ProteinPer100g *float64 `json:"protein_per_100g" example:"2.7"`
	// Optional carbohydrate grams per 100g.
	CarbsPer100g *float64 `json:"carbs_per_100g" example:"28"`
	// Optional fat grams per 100g.
	FatPer100g *float64 `json:"fat_per_100g" example:"0.3"`
}

func (r *UpdateFoodRequest) Validate() error {
	if r.Name == nil && r.BrandName == nil && r.Barcode == nil && r.KcalPer100g == nil && r.ProteinPer100g == nil && r.CarbsPer100g == nil && r.FatPer100g == nil {
		return ErrNoFieldsToUpdate
	}
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrInvalidName
	}
	if (r.KcalPer100g != nil && *r.KcalPer100g < 0) ||
		(r.ProteinPer100g != nil && *r.ProteinPer100g < 0) ||
		(r.CarbsPer100g != nil && *r.CarbsPer100g < 0) ||
		(r.FatPer100g != nil && *r.FatPer100g < 0) {
		return ErrInvalidNutrition
	}
	return nil
}

func (r *UpdateFoodRequest) ToServiceInput() service.UpdateFoodInput {
	return service.UpdateFoodInput{
		Name:           r.Name,
		BrandName:      r.BrandName,
		Barcode:        r.Barcode,
		KcalPer100g:    r.KcalPer100g,
		ProteinPer100g: r.ProteinPer100g,
		CarbsPer100g:   r.CarbsPer100g,
		FatPer100g:     r.FatPer100g,
	}
}
