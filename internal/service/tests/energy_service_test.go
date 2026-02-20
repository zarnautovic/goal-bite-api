package service_test

import (
	"context"
	"testing"
	"time"

	"goal-bite-api/internal/domain/bodyweightlog"
	"goal-bite-api/internal/domain/user"
	"goal-bite-api/internal/repository"
	"goal-bite-api/internal/service"
)

type fakeEnergyUserReader struct {
	getFn func(ctx context.Context, id uint) (user.User, error)
}

func (f fakeEnergyUserReader) GetByID(ctx context.Context, id uint) (user.User, error) {
	if f.getFn == nil {
		return user.User{}, nil
	}
	return f.getFn(ctx, id)
}

type fakeEnergyWeightReader struct {
	listFn func(ctx context.Context, userID uint, from, to time.Time) ([]bodyweightlog.BodyWeightLog, error)
}

func (f fakeEnergyWeightReader) ListByRangeAll(ctx context.Context, userID uint, from, to time.Time) ([]bodyweightlog.BodyWeightLog, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, userID, from, to)
}

type fakeEnergyTotalsReader struct {
	getFn func(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error)
}

func (f fakeEnergyTotalsReader) GetDailyTotals(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error) {
	if f.getFn == nil {
		return repository.DailyTotals{}, nil
	}
	return f.getFn(ctx, userID, date)
}

func TestEnergyServiceGetProgress(t *testing.T) {
	sex := "male"
	activity := "moderate"
	height := 180.0
	birth := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	svc := service.NewEnergyService(
		fakeEnergyUserReader{getFn: func(_ context.Context, _ uint) (user.User, error) {
			return user.User{
				ID:            1,
				Sex:           &sex,
				HeightCM:      &height,
				BirthDate:     &birth,
				ActivityLevel: &activity,
			}, nil
		}},
		fakeEnergyWeightReader{listFn: func(_ context.Context, _ uint, _, _ time.Time) ([]bodyweightlog.BodyWeightLog, error) {
			return []bodyweightlog.BodyWeightLog{
				{WeightKG: 85.0, LoggedAt: time.Date(2026, 2, 1, 8, 0, 0, 0, time.UTC)},
				{WeightKG: 84.5, LoggedAt: time.Date(2026, 2, 8, 8, 0, 0, 0, time.UTC)},
				{WeightKG: 84.0, LoggedAt: time.Date(2026, 2, 15, 8, 0, 0, 0, time.UTC)},
			}, nil
		}},
		fakeEnergyTotalsReader{getFn: func(_ context.Context, _ uint, _ time.Time) (repository.DailyTotals, error) {
			return repository.DailyTotals{Kcal: 2200}, nil
		}},
	)

	out, err := svc.GetProgress(context.Background(), service.EnergyProgressInput{
		UserID: 1,
		From:   "2026-02-01",
		To:     "2026-02-15",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out.ObservedTDEEKcal == 0 || out.RecommendedTDEEKcal == 0 {
		t.Fatalf("unexpected output: %+v", out)
	}
}
