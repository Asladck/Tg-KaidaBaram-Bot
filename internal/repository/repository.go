package repository

import (
	"github.com/jmoiron/sqlx"
	"tg-bot/internal/models"
	"time"
)

type Auth interface {
	Create(user models.User) (int64, error)
	GetUserById(id int64) (models.User, error)
}
type Stats interface {
	Save(stat models.Statistic) error
}
type Events interface {
	Exists(title string) (bool, error)
	DeleteOldEvents(now time.Time) error
	AddEvent(e models.Event) error
	GetRecentEvents(limit int) ([]models.Event, error)
}
type Repository struct {
	Auth
	Stats
	Events
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Auth:   NewAuthPostgres(db),
		Stats:  NewStatsPostgres(db),
		Events: NewEventPostgres(db),
	}
}
