package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"nutrition/internal/domain/usergoal"
	"nutrition/internal/repository"
	"nutrition/internal/service"
)

type fakeUserGoalStore struct {
	upsertFn func(ctx context.Context, in repository.UpsertUserGoalInput) (usergoal.UserGoal, error)
	getFn    func(ctx context.Context, userID uint) (usergoal.UserGoal, error)
}

func (f fakeUserGoalStore) Upsert(ctx context.Context, in repository.UpsertUserGoalInput) (usergoal.UserGoal, error) {
	if f.upsertFn == nil {
		return usergoal.UserGoal{}, nil
	}
	return f.upsertFn(ctx, in)
}

func (f fakeUserGoalStore) GetByUserID(ctx context.Context, userID uint) (usergoal.UserGoal, error) {
	if f.getFn == nil {
		return usergoal.UserGoal{}, nil
	}
	return f.getFn(ctx, userID)
}

type fakeDailyTotalsStore struct {
	getDailyTotalsFn func(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error)
}

func (f fakeDailyTotalsStore) GetDailyTotals(ctx context.Context, userID uint, date time.Time) (repository.DailyTotals, error) {
	if f.getDailyTotalsFn == nil {
		return repository.DailyTotals{}, nil
	}
	return f.getDailyTotalsFn(ctx, userID, date)
}

func TestUserGoalService(t *testing.T) {
	t.Run("upsert validates targets", func(t *testing.T) {
		svc := service.NewUserGoalService(fakeUserGoalStore{}, fakeDailyTotalsStore{})
		_, err := svc.Upsert(context.Background(), service.UpsertUserGoalInput{
			UserID:         1,
			TargetKcal:     0,
			TargetProteinG: 100,
			TargetCarbsG:   200,
			TargetFatG:     70,
		})
		if !errors.Is(err, service.ErrInvalidTargetKcal) {
			t.Fatalf("expected ErrInvalidTargetKcal, got %v", err)
		}
	})

	t.Run("get maps not found", func(t *testing.T) {
		svc := service.NewUserGoalService(fakeUserGoalStore{
			getFn: func(_ context.Context, _ uint) (usergoal.UserGoal, error) {
				return usergoal.UserGoal{}, repository.ErrNotFound
			},
		}, fakeDailyTotalsStore{})
		_, err := svc.GetByUserID(context.Background(), 1)
		if !errors.Is(err, service.ErrUserGoalNotFound) {
			t.Fatalf("expected ErrUserGoalNotFound, got %v", err)
		}
	})

	t.Run("daily progress combines totals and targets", func(t *testing.T) {
		svc := service.NewUserGoalService(fakeUserGoalStore{
			getFn: func(_ context.Context, userID uint) (usergoal.UserGoal, error) {
				if userID != 1 {
					t.Fatalf("unexpected userID %d", userID)
				}
				return usergoal.UserGoal{
					UserID:         1,
					TargetKcal:     2200,
					TargetProteinG: 150,
					TargetCarbsG:   220,
					TargetFatG:     70,
				}, nil
			},
		}, fakeDailyTotalsStore{
			getDailyTotalsFn: func(_ context.Context, userID uint, date time.Time) (repository.DailyTotals, error) {
				if userID != 1 || date.Format("2006-01-02") != "2026-02-17" {
					t.Fatalf("unexpected args user=%d date=%s", userID, date.Format("2006-01-02"))
				}
				return repository.DailyTotals{Kcal: 1800, Protein: 120, Carbs: 180, Fat: 50}, nil
			},
		})

		out, err := svc.GetDailyProgress(context.Background(), 1, "2026-02-17")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if out.RemainingKcal != 400 {
			t.Fatalf("expected remaining kcal 400, got %v", out.RemainingKcal)
		}
		if out.RemainingProteinG != 30 {
			t.Fatalf("expected remaining protein 30, got %v", out.RemainingProteinG)
		}
	})
}
