package models

import "time"

type User struct {
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	ChatID    int64     `db:"chat_id"`
	CreatedAt time.Time `db:"created_at"` // для истории
}
