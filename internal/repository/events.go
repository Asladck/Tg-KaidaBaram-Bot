package repository

import (
	"time"

	"github.com/jmoiron/sqlx"
	"tg-bot/internal/models"
)

type EventPostgres struct {
	db *sqlx.DB
}

func NewEventPostgres(db *sqlx.DB) *EventPostgres {
	return &EventPostgres{db: db}
}

func (r *EventPostgres) GetRecentEvents(limit int) ([]models.Event, error) {
	query := `SELECT id, title, date, image_url, link
	          FROM events
	          WHERE date >= NOW()
	          ORDER BY date ASC
	          LIMIT $1`
	var events []models.Event
	err := r.db.Select(&events, query, limit)
	return events, err
}

func (r *EventPostgres) AddEvent(e models.Event) error {
	query := `INSERT INTO events (title, date, image_url, link) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, e.Title, e.Date)
	return err
}

func (r *EventPostgres) Exists(title string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM events WHERE title=$1)`
	err := r.db.Get(&exists, query, title)
	return exists, err
}

func (r *EventPostgres) DeleteOldEvents(now time.Time) error {
	query := `DELETE FROM events WHERE date < $1`
	_, err := r.db.Exec(query, now)
	return err
}
