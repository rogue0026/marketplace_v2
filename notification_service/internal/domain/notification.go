package domain

import "time"

type Notification struct {
	ID        uint64
	UserID    uint64
	Title     string
	Body      string
	CreatedAt time.Time
}
