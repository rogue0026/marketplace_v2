package apperrors

import "errors"

var (
	ErrNotificationsNotFound = errors.New("notifications not found")
	ErrInvalidUserInput      = errors.New("invalid user input")
)
