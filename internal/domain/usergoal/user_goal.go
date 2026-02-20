package usergoal

import "time"

type UserGoal struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"column:user_id;not null;uniqueIndex"`
	TargetKcal     float64   `json:"target_kcal" gorm:"column:target_kcal"`
	TargetProteinG float64   `json:"target_protein_g" gorm:"column:target_protein_g"`
	TargetCarbsG   float64   `json:"target_carbs_g" gorm:"column:target_carbs_g"`
	TargetFatG     float64   `json:"target_fat_g" gorm:"column:target_fat_g"`
	WeightGoalKG   *float64  `json:"weight_goal_kg,omitempty" gorm:"column:weight_goal_kg"`
	ActivityLevel  *string   `json:"activity_level,omitempty" gorm:"column:activity_level"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

func (UserGoal) TableName() string {
	return "user_goals"
}
