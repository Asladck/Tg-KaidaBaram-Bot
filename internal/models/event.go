package models

import "time"

type Event struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Category    string    `db:"category"`
	Date        time.Time `db:"date"`
	Location    string    `db:"location"`
	Description string    `db:"description"`
	URL         string    `db:"url"`
	ImageURL    *string   `db:"image_url"`
	CreatorTgID int64     `db:"creator_telegram_id"`
	CreatorID   int64     `db:"creator_id"`
	Status      string    `db:"status"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
