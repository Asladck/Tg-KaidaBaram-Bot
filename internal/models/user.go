package models

type User struct {
	ID       int64
	TgID     int64
	Username string
	ChatID   int64
	Category []string
}
