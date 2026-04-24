package domain

type Notification struct {
	ID            uint64 `json:"id"`
	UserID        uint64 `json:"user_id"`
	Title         string `json:"title"`
	Body          string `json:"body"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}
