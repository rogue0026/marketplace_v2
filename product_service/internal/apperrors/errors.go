package apperrors

import "errors"

var (
	ErrProductDoesNotExist = errors.New("product does not exists")
	ErrInvalidUserInput    = errors.New("invalid user input")
	ErrProductNotFound     = errors.New("product not found")
	ErrNotEnoughProducts   = errors.New("not enough products")
)
