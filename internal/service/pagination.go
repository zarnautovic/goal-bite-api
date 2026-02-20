package service

const (
	DefaultLimit  = 20
	DefaultOffset = 0
	MaxLimit      = 100
)

func IsValidPagination(limit, offset int) bool {
	if limit <= 0 || limit > MaxLimit {
		return false
	}
	return offset >= 0
}
