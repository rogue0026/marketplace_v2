package errmap

import (
	"errors"
	"gateway/internal/apperrors"
	"net/http"
)

func MapError(err error) (string, int) {
	switch {
	case errors.Is(err, apperrors.ErrInvalidUserInput):
		return "invalid user input", http.StatusBadRequest

	case errors.Is(err, apperrors.ErrProductsNotFound):
		return "products not found", http.StatusNotFound

	case errors.Is(err, apperrors.ErrNotificationsNotFound):
		return "notifications not found", http.StatusNotFound

	case errors.Is(err, apperrors.ErrUserAlreadyExists):
		return "user already exists", http.StatusConflict

	case errors.Is(err, apperrors.ErrUserNotFound):
		return "user not found", http.StatusNotFound

	case errors.Is(err, apperrors.ErrNotEnoughMoney):
		return "not enough money", http.StatusBadRequest

	case errors.Is(err, apperrors.ErrNotEnoughProducts):
		return "not enough products", http.StatusBadRequest

	default:
		return err.Error(), http.StatusInternalServerError
	}
}
