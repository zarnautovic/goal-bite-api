package service

import (
	"context"
	"errors"
	"time"

	"goal-bite-api/internal/domain/usergoal"
	"goal-bite-api/internal/repository"
)

var (
	ErrUserGoalNotFound      = errors.New("user goal not found")
	ErrInvalidTargetKcal     = errors.New("invalid target kcal")
	ErrInvalidTargetProteinG = errors.New("invalid target protein")
	ErrInvalidTargetCarbsG   = errors.New("invalid target carbs")
	ErrInvalidTargetFatG     = errors.New("invalid target fat")
	ErrInvalidWeightGoalKG   = errors.New("invalid weight goal")
	ErrInvalidActivityLevel  = errors.New("invalid activity level")
	ErrInvalidProgressDate   = errors.New("invalid progress date")
)

type UserGoalStore interface {
	Upsert(ctx context.Context, in repository.UpsertUserGoalInput) (usergoal.UserGoal, error)
	GetByUserID(ctx context.Context, userID uint) (usergoal.UserGoal, error)
}

type DailyTotalsStore interface {
	GetDailyTotals(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error)
}

type UserGoalService struct {
	repo   UserGoalStore
	totals DailyTotalsStore
}

type UpsertUserGoalInput struct {
	UserID         uint
	TargetKcal     float64
	TargetProteinG float64
	TargetCarbsG   float64
	TargetFatG     float64
	WeightGoalKG   *float64
	ActivityLevel  *string
}

type DailyProgressOutput struct {
	Date              string  `json:"date"`
	TotalKcal         float64 `json:"total_kcal"`
	TotalProteinG     float64 `json:"total_protein_g"`
	TotalCarbsG       float64 `json:"total_carbs_g"`
	TotalFatG         float64 `json:"total_fat_g"`
	TargetKcal        float64 `json:"target_kcal"`
	TargetProteinG    float64 `json:"target_protein_g"`
	TargetCarbsG      float64 `json:"target_carbs_g"`
	TargetFatG        float64 `json:"target_fat_g"`
	RemainingKcal     float64 `json:"remaining_kcal"`
	RemainingProteinG float64 `json:"remaining_protein_g"`
	RemainingCarbsG   float64 `json:"remaining_carbs_g"`
	RemainingFatG     float64 `json:"remaining_fat_g"`
}

func NewUserGoalService(repo UserGoalStore, totals DailyTotalsStore) *UserGoalService {
	return &UserGoalService{repo: repo, totals: totals}
}

func (s *UserGoalService) Upsert(ctx context.Context, in UpsertUserGoalInput) (usergoal.UserGoal, error) {
	if in.UserID == 0 {
		return usergoal.UserGoal{}, ErrInvalidUserID
	}
	if in.TargetKcal <= 0 {
		return usergoal.UserGoal{}, ErrInvalidTargetKcal
	}
	if in.TargetProteinG <= 0 {
		return usergoal.UserGoal{}, ErrInvalidTargetProteinG
	}
	if in.TargetCarbsG <= 0 {
		return usergoal.UserGoal{}, ErrInvalidTargetCarbsG
	}
	if in.TargetFatG <= 0 {
		return usergoal.UserGoal{}, ErrInvalidTargetFatG
	}
	if in.WeightGoalKG != nil && *in.WeightGoalKG <= 0 {
		return usergoal.UserGoal{}, ErrInvalidWeightGoalKG
	}
	if in.ActivityLevel != nil && !isValidActivityLevel(*in.ActivityLevel) {
		return usergoal.UserGoal{}, ErrInvalidActivityLevel
	}

	return s.repo.Upsert(ctx, repository.UpsertUserGoalInput{
		UserID:         in.UserID,
		TargetKcal:     in.TargetKcal,
		TargetProteinG: in.TargetProteinG,
		TargetCarbsG:   in.TargetCarbsG,
		TargetFatG:     in.TargetFatG,
		WeightGoalKG:   in.WeightGoalKG,
		ActivityLevel:  in.ActivityLevel,
	})
}

func (s *UserGoalService) GetByUserID(ctx context.Context, userID uint) (usergoal.UserGoal, error) {
	if userID == 0 {
		return usergoal.UserGoal{}, ErrInvalidUserID
	}
	value, err := s.repo.GetByUserID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return usergoal.UserGoal{}, ErrUserGoalNotFound
	}
	if err != nil {
		return usergoal.UserGoal{}, err
	}
	return value, nil
}

func (s *UserGoalService) GetDailyProgress(ctx context.Context, userID uint, date string) (DailyProgressOutput, error) {
	if userID == 0 {
		return DailyProgressOutput{}, ErrInvalidUserID
	}
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return DailyProgressOutput{}, ErrInvalidProgressDate
	}

	goal, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return DailyProgressOutput{}, err
	}

	totals, err := s.totals.GetDailyTotals(ctx, userID, parsedDate.UTC())
	if err != nil {
		return DailyProgressOutput{}, err
	}

	return DailyProgressOutput{
		Date:              parsedDate.Format("2006-01-02"),
		TotalKcal:         totals.Kcal,
		TotalProteinG:     totals.Protein,
		TotalCarbsG:       totals.Carbs,
		TotalFatG:         totals.Fat,
		TargetKcal:        goal.TargetKcal,
		TargetProteinG:    goal.TargetProteinG,
		TargetCarbsG:      goal.TargetCarbsG,
		TargetFatG:        goal.TargetFatG,
		RemainingKcal:     goal.TargetKcal - totals.Kcal,
		RemainingProteinG: goal.TargetProteinG - totals.Protein,
		RemainingCarbsG:   goal.TargetCarbsG - totals.Carbs,
		RemainingFatG:     goal.TargetFatG - totals.Fat,
	}, nil
}

func isValidActivityLevel(value string) bool {
	switch value {
	case "sedentary", "light", "moderate", "active", "very_active":
		return true
	default:
		return false
	}
}
