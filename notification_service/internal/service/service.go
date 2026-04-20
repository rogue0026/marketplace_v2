package service

import (
	"context"
	"notification_service/internal/domain"
)

type NotificationsRepository interface {
	GetNotifications(ctx context.Context, userID uint64, limit uint64, offset uint64) ([]*domain.Notification, error)
}

type Service struct {
	notifications NotificationsRepository
}

func New(repo NotificationsRepository) *Service {
	return &Service{
		notifications: repo,
	}
}

func (s *Service) NotificationsByUser(
	ctx context.Context,
	userID uint64,
	limit uint64,
	offset uint64,
) ([]*domain.Notification, error) {
	return s.notifications.GetNotifications(ctx, userID, limit, offset)
}
