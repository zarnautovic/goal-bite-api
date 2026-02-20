package dto

import (
	"strconv"
	"strings"

	"nutrition/internal/service"
)

type PaginationQuery struct {
	Limit  int
	Offset int
}

func ParsePagination(limitRaw, offsetRaw string) (PaginationQuery, error) {
	limit := service.DefaultLimit
	offset := service.DefaultOffset

	if strings.TrimSpace(limitRaw) != "" {
		v, err := strconv.Atoi(limitRaw)
		if err != nil {
			return PaginationQuery{}, ErrInvalidPagination
		}
		limit = v
	}
	if strings.TrimSpace(offsetRaw) != "" {
		v, err := strconv.Atoi(offsetRaw)
		if err != nil {
			return PaginationQuery{}, ErrInvalidPagination
		}
		offset = v
	}

	if !service.IsValidPagination(limit, offset) {
		return PaginationQuery{}, ErrInvalidPagination
	}

	return PaginationQuery{Limit: limit, Offset: offset}, nil
}
