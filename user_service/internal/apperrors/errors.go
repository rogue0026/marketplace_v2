package apperrors

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrAlreadyExists   = errors.New("already exists")
	ErrNotFound        = errors.New("not found")
	ErrNotEnoughMoney  = errors.New("not enough money")
)
