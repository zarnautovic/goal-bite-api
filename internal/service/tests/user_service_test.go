package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goal-bite-api/internal/domain/user"
	"goal-bite-api/internal/repository"
	"goal-bite-api/internal/service"
)

type fakeUserReader struct {
	result   user.User
	err      error
	updateFn func(ctx context.Context, id uint, updates repository.UserUpdate) (user.User, error)
}

func (f fakeUserReader) GetByID(_ context.Context, _ uint) (user.User, error) {
	return f.result, f.err
}

func (f fakeUserReader) Update(ctx context.Context, id uint, updates repository.UserUpdate) (user.User, error) {
	if f.updateFn == nil {
		return f.result, f.err
	}
	return f.updateFn(ctx, id, updates)
}

func TestUserServiceGetByID(t *testing.T) {
	t.Run("returns user when repository returns user", func(t *testing.T) {
		now := time.Now().UTC()
		expected := user.User{ID: 1, Name: "Alice", CreatedAt: now, UpdatedAt: now}
		svc := service.NewUserService(fakeUserReader{result: expected})

		got, err := svc.GetByID(context.Background(), 1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.ID != expected.ID || got.Name != expected.Name {
			t.Fatalf("unexpected user: got %+v expected %+v", got, expected)
		}
	})

	t.Run("maps repository not found to service error", func(t *testing.T) {
		svc := service.NewUserService(fakeUserReader{err: repository.ErrNotFound})

		_, err := svc.GetByID(context.Background(), 42)
		if !errors.Is(err, service.ErrUserNotFound) {
			t.Fatalf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("propagates unexpected repository errors", func(t *testing.T) {
		repoErr := errors.New("database offline")
		svc := service.NewUserService(fakeUserReader{err: repoErr})

		_, err := svc.GetByID(context.Background(), 42)
		if !errors.Is(err, repoErr) {
			t.Fatalf("expected %v, got %v", repoErr, err)
		}
	})
}

func TestUserServiceUpdate(t *testing.T) {
	t.Run("validates update payload", func(t *testing.T) {
		svc := service.NewUserService(fakeUserReader{})
		_, err := svc.Update(context.Background(), 1, service.UpdateUserInput{})
		if !errors.Is(err, service.ErrNoFieldsToUpdate) {
			t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
		}
	})

	t.Run("allows clearing optional profile field with null", func(t *testing.T) {
		svc := service.NewUserService(fakeUserReader{
			updateFn: func(_ context.Context, _ uint, updates repository.UserUpdate) (user.User, error) {
				if !updates.ActivityLevelSet {
					t.Fatalf("expected ActivityLevelSet=true")
				}
				if updates.ActivityLevel != nil {
					t.Fatalf("expected activity level to be nil for clear")
				}
				return user.User{ID: 1, Name: "A", Email: "a@example.com"}, nil
			},
		})
		_, err := svc.Update(context.Background(), 1, service.UpdateUserInput{
			ActivityLevelSet: true,
			ActivityLevel:    nil,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}
