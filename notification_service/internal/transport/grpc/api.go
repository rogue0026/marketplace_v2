package grpc

import (
	"context"
	"notification_service/internal/service"
	"notification_service/internal/transport/grpc/errmap"

	"github.com/rogue0026/marketplace-proto_v2/gen/notification_service/pb"
)

type Handler struct {
	NotificationService *service.Service
	pb.UnimplementedNotificationServiceServer
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{
		NotificationService: s,
	}
}

func (h *Handler) GetUserNotifications(
	ctx context.Context,
	in *pb.GetUserNotificationsRequest,
) (*pb.GetUserNotificationsResponse, error) {
	notifications, err := h.NotificationService.NotificationsByUser(
		ctx,
		in.UserId,
		in.Limit,
		in.Offset,
	)
	if err != nil {
		return nil, errmap.MapError(err)
	}

	resp := &pb.GetUserNotificationsResponse{
		Notifications: make([]*pb.GetUserNotificationsResponse_Notification, 0, len(notifications)),
	}

	for _, n := range notifications {
		resp.Notifications = append(resp.Notifications, &pb.GetUserNotificationsResponse_Notification{
			Id:            n.ID,
			UserId:        n.UserID,
			Title:         n.Title,
			Body:          n.Body,
			CreatedAtUnix: n.CreatedAt.Unix(),
		})
	}

	return resp, nil
}
