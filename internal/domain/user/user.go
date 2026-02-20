package user

import "time"

type User struct {
	ID            uint       `json:"id" gorm:"primaryKey"`
	Name          string     `json:"name"`
	Email         string     `json:"email" gorm:"uniqueIndex"`
	Sex           *string    `json:"sex,omitempty" gorm:"column:sex"`
	BirthDate     *time.Time `json:"birth_date,omitempty" gorm:"column:birth_date"`
	HeightCM      *float64   `json:"height_cm,omitempty" gorm:"column:height_cm"`
	ActivityLevel *string    `json:"activity_level,omitempty" gorm:"column:activity_level"`
	PasswordHash  string     `json:"-" gorm:"column:password_hash"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
