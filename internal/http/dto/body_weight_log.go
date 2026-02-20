package dto

import (
	"errors"
	"time"

	"nutrition/internal/service"
)

var (
	ErrInvalidWeightKG = errors.New("invalid weight_kg")
)

type CreateBodyWeightLogRequest struct {
	// Body weight in kilograms.
	WeightKG float64 `json:"weight_kg" example:"85.2"`
	// Measurement timestamp in RFC3339 UTC.
	LoggedAt string `json:"logged_at" example:"2026-02-17T08:00:00Z"`
}

func (r *CreateBodyWeightLogRequest) Validate() error {
	if r.WeightKG <= 0 {
		return ErrInvalidWeightKG
	}
	if _, err := time.Parse(time.RFC3339, r.LoggedAt); err != nil {
		return ErrInvalidEatenAt
	}
	return nil
}

func (r *CreateBodyWeightLogRequest) ToServiceInput(userID uint) (service.CreateBodyWeightLogInput, error) {
	loggedAt, err := time.Parse(time.RFC3339, r.LoggedAt)
	if err != nil {
		return service.CreateBodyWeightLogInput{}, err
	}
	return service.CreateBodyWeightLogInput{UserID: userID, WeightKG: r.WeightKG, LoggedAt: loggedAt.UTC()}, nil
}

type ListBodyWeightLogsQuery struct {
	From   string
	To     string
	Limit  int
	Offset int
}

func (q *ListBodyWeightLogsQuery) Validate() error {
	if _, err := time.Parse("2006-01-02", q.From); err != nil {
		return ErrInvalidDateRange
	}
	if _, err := time.Parse("2006-01-02", q.To); err != nil {
		return ErrInvalidDateRange
	}
	if !service.IsValidPagination(q.Limit, q.Offset) {
		return ErrInvalidPagination
	}
	return nil
}

func (q *ListBodyWeightLogsQuery) ToServiceInput(userID uint) service.ListBodyWeightLogsInput {
	return service.ListBodyWeightLogsInput{UserID: userID, From: q.From, To: q.To, Limit: q.Limit, Offset: q.Offset}
}
