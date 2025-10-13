package models

import "time"

type User struct {
	ID        int64     `db:"id"`
	TgID      int64     `db:"telegram_id"`
	Username  string    `db:"username"`
	ChatID    int64     `db:"chat_id"`
	CreatedAt time.Time `db:"created_at"` // для истории
}
