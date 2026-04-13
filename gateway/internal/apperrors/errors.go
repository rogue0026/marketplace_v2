package apperrors

import "errors"

var (
	ErrFailedPrecondition = errors.New("failed preconditions")
	ErrAlreadyExists      = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
)
