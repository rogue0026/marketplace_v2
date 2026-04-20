package errmap

import (
	"errors"
	"notification_service/internal/apperrors"

	"github.com/rogue0026/marketplace-proto_v2/gen/notification_service/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MapError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrNotificationsNotFound):
		st := status.New(codes.NotFound, "notifications not found")
		withDetails, detailsErr := st.WithDetails(&errdetails.ErrorInfo{
			Reason: pb.Reason_NOTIFICATIONS_NOT_FOUND.String(),
			Domain: "notification.service",
		})
		if detailsErr != nil {
			return st.Err()
		}
		return withDetails.Err()

	case errors.Is(err, apperrors.ErrInvalidUserInput):
		return status.Error(codes.InvalidArgument, "invalid user input")

	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
