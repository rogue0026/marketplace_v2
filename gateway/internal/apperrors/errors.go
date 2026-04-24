package apperrors

import "errors"

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrInvalidUserInput      = errors.New("invalid user input")
	ErrProductsNotFound      = errors.New("products not found")
	ErrNotificationsNotFound = errors.New("notifications not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserNotFound          = errors.New("user not found")
	ErrNotEnoughMoney        = errors.New("not enough money")
	ErrNotEnoughProducts     = errors.New("not enough products")
	ErrBasketIsEmpty         = errors.New("basket is empty")
)
