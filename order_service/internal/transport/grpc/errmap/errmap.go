package errmap

import (
	"errors"
	"order_service/internal/apperrors"

	"github.com/rogue0026/marketplace-proto_v2/gen/order_service/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapError(err error) error {
	switch {

	case errors.Is(err, apperrors.ErrNotEnoughProducts):
		st := status.New(codes.FailedPrecondition, "not enough products")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_NOT_ENOUGH_PRODUCTS.String(),
			Domain: "order.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrNotEnoughMoney):
		st := status.New(codes.FailedPrecondition, "not enough money")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_NOT_ENOUGH_MONEY.String(),
			Domain: "order.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrProductDoesNotExists):
		st := status.New(codes.Internal, "product does not exists")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_PRODUCT_DOES_NOT_EXISTS.String(),
			Domain: "order.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrUserNotFound):
		st := status.New(codes.NotFound, "user not found")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_USER_NOT_FOUND.String(),
			Domain: "order.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrOrderNotFound):
		st := status.New(codes.NotFound, "order not found")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_ORDER_NOT_FOUND.String(),
			Domain: "order.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrEmptyBasket):
		st := status.New(codes.FailedPrecondition, "unable to create order. basket is empty")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_BASKET_IS_EMPTY.String(),
			Domain: "order.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
