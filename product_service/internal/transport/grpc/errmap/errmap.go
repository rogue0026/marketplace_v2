package errmap

import (
	"errors"
	"product_service/internal/apperrors"

	"github.com/rogue0026/marketplace-proto_v2/gen/product_service/pb"
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
			Domain: "product.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrProductDoesNotExist):
		st := status.New(codes.NotFound, "requested product does not exists")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_PRODUCTS_NOT_FOUND.String(),
			Domain: "product.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrInvalidUserInput):
		st := status.New(codes.InvalidArgument, "invalid user input data")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_INVALID_USER_INPUT.String(),
			Domain: "product.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrProductNotFound):
		st := status.New(codes.NotFound, "requested product not found")
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_PRODUCTS_NOT_FOUND.String(),
			Domain: "product.service",
		})
		if err != nil {
			return st.Err()
		}
		return withDetails.Err()

	default:
		st := status.New(codes.Unknown, err.Error())
		withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
			Reason: "UNKNOWN_REASON",
			Domain: "product.service",
		})
		if err != nil {
			return st.Err()
		}

		return withDetails.Err()
	}
}
