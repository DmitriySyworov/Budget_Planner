package handler_request

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/go-playground/validator/v10"
)

var (
	ErrIncorrectFormatBody = errors.New("incorrect format body")
)

func HandlerRequest[T any](body io.ReadCloser) (*T, error) {
	var payload T
	errDecode := json.NewDecoder(body).Decode(&payload)
	if errDecode != nil {
		return nil, ErrIncorrectFormatBody
	}
	errValidate := validator.New().Struct(&payload)
	if errValidate != nil {
		return &payload, errValidate
	}
	return &payload, nil
}
