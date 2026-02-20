package handlers

import (
	"net/http"

	"nutrition/internal/http/dto"
)

func parsePagination(r *http.Request) (int, int, bool) {
	p, err := dto.ParsePagination(r.URL.Query().Get("limit"), r.URL.Query().Get("offset"))
	if err != nil {
		return 0, 0, false
	}
	return p.Limit, p.Offset, true
}
