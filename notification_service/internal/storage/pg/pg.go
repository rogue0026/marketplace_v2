package pg

import (
	"context"
	"fmt"
	"notification_service/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type NotificationsRepository struct {
	pool *pgxpool.Pool
}

func NewNotificationsRepository(pool *pgxpool.Pool) *NotificationsRepository {
	return &NotificationsRepository{
		pool: pool,
	}
}

/*
CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);
*/

const GetNotificationsSQL = `
	SELECT id, user_id, title, body, created_at
	FROM notifications
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
`

const CreateNotificationSQL = `
	INSERT INTO notifications (user_id, title, body, created_at)
	VALUES ($1, $2, $3, $4)
`

func (r *NotificationsRepository) CreateNotification(ctx context.Context, item *domain.Notification) error {
	_, err := r.pool.Exec(
		ctx,
		CreateNotificationSQL,
		item.UserID,
		item.Title,
		item.Body,
		item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("repo, create notification: %w", err)
	}

	return nil
}

func (r *NotificationsRepository) GetNotifications(
	ctx context.Context,
	userID uint64,
	limit uint64,
	offset uint64,
) ([]*domain.Notification, error) {
	rows, err := r.pool.Query(ctx, GetNotificationsSQL, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repo, get notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]*domain.Notification, 0)
	for rows.Next() {
		item := domain.Notification{}
		err = rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Title,
			&item.Body,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repo, get notifications, scan row: %w", err)
		}

		notifications = append(notifications, &item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("repo, get notifications: %w", err)
	}

	return notifications, nil
}
