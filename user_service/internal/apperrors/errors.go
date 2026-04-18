package apperrors

import "errors"

var (
	ErrProductsNotFound     = errors.New("products not found")
	ErrEmptyBasket          = errors.New("no data")
	ErrNotEnoughMoney       = errors.New("not enough money")
	ErrUsernameAlreadyTaken = errors.New("username already taken")
	ErrUserNotFound         = errors.New("user not found")
)
