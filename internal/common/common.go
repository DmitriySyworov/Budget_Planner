package common

import (
	"errors"
	"strconv"
)

const (
	TypeSoftDelete = "soft-delete"
	TypeHardDelete = "hard-delete"

	maxLimit      = 100
	defaultLimit  = 50
	defaultOffset = 0
)

var (
	ErrIncorrectLimit  = errors.New("the limit must be a positive integer not greater than 100")
	ErrIncorrectOffset = errors.New("the offset must be a positive integer")
)

func PaginationHelper(limitStr, offsetStr string) (int, int, []string) {
	sliceError := make([]string, 0, 2)
	var limit, offset int
	var errLimit, errOffset error
	if limitStr != "" {
		limit, errLimit = strconv.Atoi(limitStr)
		if errLimit != nil {
			sliceError = append(sliceError, ErrIncorrectLimit.Error())
		}
	} else {
		limit = defaultLimit
	}
	if limit > 100 {
		limit = maxLimit
	} else if limit < 0 {
		sliceError = append(sliceError, ErrIncorrectLimit.Error())
	}
	if offsetStr != "" {
		offset, errOffset = strconv.Atoi(offsetStr)
		if errOffset != nil {
			sliceError = append(sliceError, ErrIncorrectOffset.Error())
		}
	} else {
		offset = defaultOffset
	}
	if offset < 0 {
		sliceError = append(sliceError, ErrIncorrectOffset.Error())
	}
	return limit, offset, sliceError
}
