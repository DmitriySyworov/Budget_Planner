package shared_common

import (
	"shared/shared_errors"
	"strconv"
)

var (
	TypeSoftDelete = "soft-delete"
	TypeHardDelete = "hard-delete"
)

const (
	maxLimit      = 100
	defaultLimit  = 50
	defaultOffset = 0
)

func PaginationHelper(limitStr, offsetStr string) (int, int, []error) {
	sliceError := make([]error, 0, 2)
	var limit, offset int
	var errLimit, errOffset error
	if limitStr != "" {
		limit, errLimit = strconv.Atoi(limitStr)
		if errLimit != nil {
			sliceError = append(sliceError, shared_errors.ErrIncorrectLimit)
		}
	} else {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = maxLimit
	} else if limit < 0 {
		sliceError = append(sliceError, shared_errors.ErrIncorrectLimit)
	}
	if offsetStr != "" {
		offset, errOffset = strconv.Atoi(offsetStr)
		if errOffset != nil {
			sliceError = append(sliceError, shared_errors.ErrIncorrectOffset)
		}
	} else {
		offset = defaultOffset
	}
	if offset < 0 {
		sliceError = append(sliceError, shared_errors.ErrIncorrectOffset)
	}
	return limit, offset, sliceError
}
