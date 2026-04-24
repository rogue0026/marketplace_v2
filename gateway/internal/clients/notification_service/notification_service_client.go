package notification_service

import (
	"context"
	"errors"
	"fmt"
	"gateway/internal/apperrors"
	"gateway/internal/domain"

	"github.com/rogue0026/marketplace-proto_v2/gen/notification_service/pb"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type errKey struct {
	code   codes.Code
	reason string
}

var errorsMap = map[errKey]error{
	{
		code:   codes.NotFound,
		reason: pb.Reason_NOTIFICATIONS_NOT_FOUND.String(),
	}: apperrors.ErrNotificationsNotFound,
}

var ErrEmptyReason = errors.New("reason is empty")

func extractReason(s *status.Status) (string, error) {
	for _, d := range s.Details() {
		errInfo, ok := d.(*errdetails.ErrorInfo)
		if ok {
			return errInfo.Reason, nil
		}
	}

	return "", ErrEmptyReason
}

func mapErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	reason, reasonErr := extractReason(st)
	if reasonErr == nil {
		k := errKey{
			code:   st.Code(),
			reason: reason,
		}
		if appErr, exists := errorsMap[k]; exists {
			return appErr
		}
	}

	if st.Code() == codes.NotFound {
		return apperrors.ErrNotificationsNotFound
	}

	if st.Code() == codes.InvalidArgument {
		return apperrors.ErrInvalidUserInput
	}

	return err
}

type NotificationService struct {
	client pb.NotificationServiceClient
}

func NewNotificationService(ccAddr string) (*NotificationService, error) {
	cc, err := grpc.NewClient(ccAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("client, notification service: %w", err)
	}

	client := pb.NewNotificationServiceClient(cc)
	s := &NotificationService{
		client: client,
	}

	return s, nil
}

func (s *NotificationService) NotificationsByUser(
	ctx context.Context,
	userID uint64,
	limit uint64,
	offset uint64,
) ([]*domain.Notification, error) {
	resp, err := s.client.GetUserNotifications(ctx, &pb.GetUserNotificationsRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("client, notification service, get user notifications: %w", mapErr(err))
	}

	notifications := make([]*domain.Notification, 0, len(resp.Notifications))
	for _, n := range resp.Notifications {
		notifications = append(notifications, &domain.Notification{
			ID:            n.Id,
			UserID:        n.UserId,
			Title:         n.Title,
			Body:          n.Body,
			CreatedAtUnix: n.CreatedAtUnix,
		})
	}

	return notifications, nil
}
