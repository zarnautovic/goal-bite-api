package food

import "time"

type Food struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"column:user_id;not null"`
	Name           string    `json:"name"`
	BrandName      *string   `json:"brand_name,omitempty" gorm:"column:brand_name"`
	Barcode        *string   `json:"barcode,omitempty" gorm:"column:barcode"`
	KcalPer100g    float64   `json:"kcal_per_100g" gorm:"column:kcal_per_100g"`
	ProteinPer100g float64   `json:"protein_per_100g" gorm:"column:protein_per_100g"`
	CarbsPer100g   float64   `json:"carbs_per_100g" gorm:"column:carbs_per_100g"`
	FatPer100g     float64   `json:"fat_per_100g" gorm:"column:fat_per_100g"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
