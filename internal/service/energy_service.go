package service

import (
	"context"
	"errors"
	"math"
	"time"

	"goal-bite-api/internal/domain/bodyweightlog"
	"goal-bite-api/internal/domain/user"
	"goal-bite-api/internal/repository"
)

var (
	ErrInvalidEnergyProgressQuery = errors.New("invalid energy progress query")
	ErrInsufficientWeightData     = errors.New("insufficient weight data")
	ErrInsufficientIntakeData     = errors.New("insufficient intake data")
)

type EnergyUserReader interface {
	GetByID(ctx context.Context, id uint) (user.User, error)
}

type EnergyWeightReader interface {
	ListByRangeAll(ctx context.Context, userID uint, from, to time.Time) ([]bodyweightlog.BodyWeightLog, error)
}

type EnergyTotalsReader interface {
	GetDailyTotals(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error)
}

type EnergyService struct {
	users   EnergyUserReader
	weights EnergyWeightReader
	totals  EnergyTotalsReader
}

type EnergyProgressInput struct {
	UserID uint
	From   string
	To     string
}

type EnergyProgressOutput struct {
	From                 string   `json:"from"`
	To                   string   `json:"to"`
	AvgIntakeKcal        float64  `json:"avg_intake_kcal"`
	WeightTrendKgPerWeek float64  `json:"weight_trend_kg_per_week"`
	ObservedTDEEKcal     float64  `json:"observed_tdee_kcal"`
	FormulaTDEEKcal      *float64 `json:"formula_tdee_kcal,omitempty"`
	RecommendedTDEEKcal  float64  `json:"recommended_tdee_kcal"`
	DataQualityScore     float64  `json:"data_quality_score"`
}

func NewEnergyService(users EnergyUserReader, weights EnergyWeightReader, totals EnergyTotalsReader) *EnergyService {
	return &EnergyService{users: users, weights: weights, totals: totals}
}

func (s *EnergyService) GetProgress(ctx context.Context, in EnergyProgressInput) (EnergyProgressOutput, error) {
	if in.UserID == 0 {
		return EnergyProgressOutput{}, ErrInvalidUserID
	}
	fromDate, err := time.Parse("2006-01-02", in.From)
	if err != nil {
		return EnergyProgressOutput{}, ErrInvalidEnergyProgressQuery
	}
	toDate, err := time.Parse("2006-01-02", in.To)
	if err != nil {
		return EnergyProgressOutput{}, ErrInvalidEnergyProgressQuery
	}
	if toDate.Before(fromDate) {
		return EnergyProgressOutput{}, ErrInvalidEnergyProgressQuery
	}
	if int(toDate.Sub(fromDate).Hours()/24) > 90 {
		return EnergyProgressOutput{}, ErrInvalidEnergyProgressQuery
	}

	rangeStart := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
	rangeEndExclusive := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)
	logs, err := s.weights.ListByRangeAll(ctx, in.UserID, rangeStart, rangeEndExclusive)
	if err != nil {
		return EnergyProgressOutput{}, err
	}
	if len(logs) < 2 {
		return EnergyProgressOutput{}, ErrInsufficientWeightData
	}

	trendKgPerDay := calculateWeightTrendKgPerDay(logs)
	trendKgPerWeek := trendKgPerDay * 7

	totalDays := int(toDate.Sub(fromDate).Hours()/24) + 1
	var sumIntake float64
	var intakeDays int
	for day := 0; day < totalDays; day++ {
		date := rangeStart.Add(time.Duration(day) * 24 * time.Hour)
		t, err := s.totals.GetDailyTotals(ctx, in.UserID, date)
		if err != nil {
			return EnergyProgressOutput{}, err
		}
		if t.Kcal > 0 {
			sumIntake += t.Kcal
			intakeDays++
		}
	}
	if intakeDays == 0 {
		return EnergyProgressOutput{}, ErrInsufficientIntakeData
	}
	avgIntake := sumIntake / float64(intakeDays)
	observedTDEE := avgIntake - trendKgPerDay*7700

	dataQuality := math.Min(float64(len(logs))/10.0, 1.0)*0.5 + (float64(intakeDays)/float64(totalDays))*0.5

	var formulaTDEE *float64
	recommended := observedTDEE
	u, err := s.users.GetByID(ctx, in.UserID)
	if err == nil {
		if v, ok := calculateFormulaTDEE(u, logs[len(logs)-1].WeightKG, toDate); ok {
			formulaTDEE = &v
			recommended = v*(1-dataQuality) + observedTDEE*dataQuality
		}
	} else if !errors.Is(err, ErrUserNotFound) && !errors.Is(err, repository.ErrNotFound) {
		return EnergyProgressOutput{}, err
	}

	return EnergyProgressOutput{
		From:                 fromDate.Format("2006-01-02"),
		To:                   toDate.Format("2006-01-02"),
		AvgIntakeKcal:        round2(avgIntake),
		WeightTrendKgPerWeek: round3(trendKgPerWeek),
		ObservedTDEEKcal:     round2(observedTDEE),
		FormulaTDEEKcal:      formulaTDEE,
		RecommendedTDEEKcal:  round2(recommended),
		DataQualityScore:     round3(dataQuality),
	}, nil
}

func calculateWeightTrendKgPerDay(logs []bodyweightlog.BodyWeightLog) float64 {
	if len(logs) < 2 {
		return 0
	}
	base := logs[0].LoggedAt.UTC()
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(logs))
	for _, l := range logs {
		x := l.LoggedAt.UTC().Sub(base).Hours() / 24.0
		y := l.WeightKG
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

func calculateFormulaTDEE(u user.User, weightKG float64, referenceDate time.Time) (float64, bool) {
	if u.Sex == nil || u.BirthDate == nil || u.HeightCM == nil || u.ActivityLevel == nil {
		return 0, false
	}
	age := referenceDate.Year() - u.BirthDate.Year()
	birthdayThisYear := time.Date(referenceDate.Year(), u.BirthDate.Month(), u.BirthDate.Day(), 0, 0, 0, 0, time.UTC)
	if referenceDate.Before(birthdayThisYear) {
		age--
	}
	if age <= 0 {
		return 0, false
	}

	var bmr float64
	switch *u.Sex {
	case "male":
		bmr = 10*weightKG + 6.25*(*u.HeightCM) - 5*float64(age) + 5
	case "female":
		bmr = 10*weightKG + 6.25*(*u.HeightCM) - 5*float64(age) - 161
	default:
		return 0, false
	}

	var multiplier float64
	switch *u.ActivityLevel {
	case "sedentary":
		multiplier = 1.2
	case "light":
		multiplier = 1.375
	case "moderate":
		multiplier = 1.55
	case "active":
		multiplier = 1.725
	case "very_active":
		multiplier = 1.9
	default:
		return 0, false
	}

	return bmr * multiplier, true
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}
