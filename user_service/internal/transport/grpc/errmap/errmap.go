package errmap

import (
	"errors"
	"github.com/rogue0026/marketplace-proto_v2/gen/user_service/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"user_service/internal/apperrors"
)

func MapError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrProductsNotFound):
		st := status.New(codes.NotFound, "products not found")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_PRODUCT_NOT_FOUND.String(),
			Domain: "user.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrEmptyBasket):
		st := status.New(codes.NotFound, "basket is empty")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_BASKET_IS_EMPTY.String(),
			Domain: "user.service",
		})
		if err != nil {
			return st.Err()
		}

		return withDetails.Err()

	case errors.Is(err, apperrors.ErrNotEnoughMoney):
		st := status.New(codes.FailedPrecondition, "not enough money")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_NOT_ENOUGH_MONEY.String(),
			Domain: "user.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrUserNotFound):
		st := status.New(codes.NotFound, "user not found")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_USER_NOT_FOUND.String(),
			Domain: "user.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrUsernameAlreadyTaken):
		st := status.New(codes.AlreadyExists, "username already exists")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_USER_ALREADY_EXISTS.String(),
			Domain: "user.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
