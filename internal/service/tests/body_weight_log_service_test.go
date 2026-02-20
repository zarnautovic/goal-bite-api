package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goal-bite-api/internal/domain/bodyweightlog"
	"goal-bite-api/internal/repository"
	"goal-bite-api/internal/service"
)

type fakeBodyWeightLogStore struct {
	createFn func(ctx context.Context, in repository.CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error)
	listFn   func(ctx context.Context, userID uint, from, to time.Time, limit, offset int) ([]bodyweightlog.BodyWeightLog, error)
	latestFn func(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error)
}

func (f fakeBodyWeightLogStore) Create(ctx context.Context, in repository.CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error) {
	if f.createFn == nil {
		return bodyweightlog.BodyWeightLog{}, nil
	}
	return f.createFn(ctx, in)
}

func (f fakeBodyWeightLogStore) ListByRange(ctx context.Context, userID uint, from, to time.Time, limit, offset int) ([]bodyweightlog.BodyWeightLog, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, userID, from, to, limit, offset)
}

func (f fakeBodyWeightLogStore) GetLatest(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error) {
	if f.latestFn == nil {
		return bodyweightlog.BodyWeightLog{}, nil
	}
	return f.latestFn(ctx, userID)
}

func TestBodyWeightLogService(t *testing.T) {
	t.Run("create validates weight", func(t *testing.T) {
		svc := service.NewBodyWeightLogService(fakeBodyWeightLogStore{})
		_, err := svc.Create(context.Background(), service.CreateBodyWeightLogInput{UserID: 1, WeightKG: 0, LoggedAt: time.Now().UTC()})
		if !errors.Is(err, service.ErrInvalidWeightKG) {
			t.Fatalf("expected ErrInvalidWeightKG, got %v", err)
		}
	})

	t.Run("latest maps not found", func(t *testing.T) {
		svc := service.NewBodyWeightLogService(fakeBodyWeightLogStore{latestFn: func(_ context.Context, _ uint) (bodyweightlog.BodyWeightLog, error) {
			return bodyweightlog.BodyWeightLog{}, repository.ErrNotFound
		}})
		_, err := svc.GetLatest(context.Background(), 1)
		if !errors.Is(err, service.ErrBodyWeightLogNotFound) {
			t.Fatalf("expected ErrBodyWeightLogNotFound, got %v", err)
		}
	})

	t.Run("list parses range", func(t *testing.T) {
		svc := service.NewBodyWeightLogService(fakeBodyWeightLogStore{listFn: func(_ context.Context, userID uint, from, to time.Time, limit, offset int) ([]bodyweightlog.BodyWeightLog, error) {
			if userID != 1 || limit != 20 || offset != 0 || !to.After(from) {
				t.Fatalf("unexpected args")
			}
			return []bodyweightlog.BodyWeightLog{{ID: 1, UserID: 1, WeightKG: 85.2, LoggedAt: from}}, nil
		}})
		values, err := svc.List(context.Background(), service.ListBodyWeightLogsInput{UserID: 1, From: "2026-02-01", To: "2026-02-17", Limit: 20, Offset: 0})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(values) != 1 {
			t.Fatalf("expected one value, got %d", len(values))
		}
	})
}
