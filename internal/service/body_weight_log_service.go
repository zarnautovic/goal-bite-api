package service

import (
	"context"
	"errors"
	"time"

	"goal-bite-api/internal/domain/bodyweightlog"
	"goal-bite-api/internal/repository"
)

var (
	ErrBodyWeightLogNotFound = errors.New("body weight log not found")
	ErrInvalidWeightKG       = errors.New("invalid weight_kg")
	ErrInvalidDateRange      = errors.New("invalid date range")
)

type BodyWeightLogStore interface {
	Create(ctx context.Context, in repository.CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error)
	ListByRange(ctx context.Context, userID uint, from, to time.Time, limit, offset int) ([]bodyweightlog.BodyWeightLog, error)
	GetLatest(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error)
}

type BodyWeightLogService struct {
	repo BodyWeightLogStore
}

type CreateBodyWeightLogInput struct {
	UserID   uint
	WeightKG float64
	LoggedAt time.Time
}

type ListBodyWeightLogsInput struct {
	UserID uint
	From   string
	To     string
	Limit  int
	Offset int
}

func NewBodyWeightLogService(repo BodyWeightLogStore) *BodyWeightLogService {
	return &BodyWeightLogService{repo: repo}
}

func (s *BodyWeightLogService) Create(ctx context.Context, in CreateBodyWeightLogInput) (bodyweightlog.BodyWeightLog, error) {
	if in.UserID == 0 {
		return bodyweightlog.BodyWeightLog{}, ErrInvalidUserID
	}
	if in.WeightKG <= 0 {
		return bodyweightlog.BodyWeightLog{}, ErrInvalidWeightKG
	}
	if in.LoggedAt.IsZero() {
		return bodyweightlog.BodyWeightLog{}, ErrInvalidEatenAt
	}

	return s.repo.Create(ctx, repository.CreateBodyWeightLogInput{
		UserID:   in.UserID,
		WeightKG: in.WeightKG,
		LoggedAt: in.LoggedAt.UTC(),
	})
}

func (s *BodyWeightLogService) List(ctx context.Context, in ListBodyWeightLogsInput) ([]bodyweightlog.BodyWeightLog, error) {
	if in.UserID == 0 {
		return nil, ErrInvalidUserID
	}
	if !IsValidPagination(in.Limit, in.Offset) {
		return nil, ErrInvalidPagination
	}

	fromDate, err := time.Parse("2006-01-02", in.From)
	if err != nil {
		return nil, ErrInvalidDateRange
	}
	toDate, err := time.Parse("2006-01-02", in.To)
	if err != nil {
		return nil, ErrInvalidDateRange
	}
	if toDate.Before(fromDate) {
		return nil, ErrInvalidDateRange
	}

	from := time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(), 0, 0, 0, 0, time.UTC)
	to := time.Date(toDate.Year(), toDate.Month(), toDate.Day(), 0, 0, 0, 0, time.UTC).Add(24 * time.Hour)

	return s.repo.ListByRange(ctx, in.UserID, from, to, in.Limit, in.Offset)
}

func (s *BodyWeightLogService) GetLatest(ctx context.Context, userID uint) (bodyweightlog.BodyWeightLog, error) {
	if userID == 0 {
		return bodyweightlog.BodyWeightLog{}, ErrInvalidUserID
	}

	value, err := s.repo.GetLatest(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return bodyweightlog.BodyWeightLog{}, ErrBodyWeightLogNotFound
	}
	if err != nil {
		return bodyweightlog.BodyWeightLog{}, err
	}
	return value, nil
}
