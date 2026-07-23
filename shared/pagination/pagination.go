package pagination

import (
	"shared/shared_errors"
	"strconv"
)

const (
	MaxLimit      = 100
	DefaultLimit  = 10
	DefaultOffset = 0
)

func HelperPagination(limitStr, offsetStr string, mapError *shared_errors.MapError) (int, int) {
	var limit, offset int
	var errLimit, errOffset error
	if limitStr != "" {
		limit, errLimit = strconv.Atoi(limitStr)
		if errLimit != nil {
			mapError.Map["limit"] = shared_errors.ErrIncorrectLimit.Error()
		}
	} else {
		limit = DefaultLimit
	}
	if limit > MaxLimit {
		mapError.Map["limit"] = shared_errors.ErrIncorrectLimit.Error()
	} else if limit <= 0 {
		mapError.Map["limit"] = shared_errors.ErrIncorrectLimit.Error()
	}
	if offsetStr != "" {
		offset, errOffset = strconv.Atoi(offsetStr)
		if errOffset != nil {
			mapError.Map["offset"] = shared_errors.ErrIncorrectOffset.Error()
		}
	} else {
		offset = DefaultOffset
	}
	if offset < 0 {
		mapError.Map["offset"] = shared_errors.ErrIncorrectOffset.Error()
	}
	return limit, offset
}
