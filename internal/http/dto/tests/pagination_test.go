package dto_test

import (
	"errors"
	"testing"

	"goal-bite-api/internal/http/dto"
	"goal-bite-api/internal/service"
)

func TestParsePagination(t *testing.T) {
	t.Run("defaults when empty", func(t *testing.T) {
		got, err := dto.ParsePagination("", "")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got.Limit != service.DefaultLimit || got.Offset != service.DefaultOffset {
			t.Fatalf("unexpected defaults: %+v", got)
		}
	})

	t.Run("rejects non-numeric", func(t *testing.T) {
		_, err := dto.ParsePagination("abc", "0")
		if !errors.Is(err, dto.ErrInvalidPagination) {
			t.Fatalf("expected ErrInvalidPagination, got %v", err)
		}
	})

	t.Run("rejects above max", func(t *testing.T) {
		_, err := dto.ParsePagination("101", "0")
		if !errors.Is(err, dto.ErrInvalidPagination) {
			t.Fatalf("expected ErrInvalidPagination, got %v", err)
		}
	})
}
