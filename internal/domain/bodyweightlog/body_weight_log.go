package bodyweightlog

import "time"

type BodyWeightLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"column:user_id"`
	WeightKG  float64   `json:"weight_kg" gorm:"column:weight_kg"`
	LoggedAt  time.Time `json:"logged_at" gorm:"column:logged_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
