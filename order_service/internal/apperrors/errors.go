package apperrors

import "errors"

var (
	ErrProductDoesNotExists = errors.New("product does not exists")
	ErrOrderNotFound        = errors.New("order not found")
	ErrEmptyBasket          = errors.New("basket is empty")
	ErrUserNotFound         = errors.New("user not found")
	ErrNotEnoughMoney       = errors.New("not enough money")
	ErrNotEnoughProducts    = errors.New("not enough products")
)
