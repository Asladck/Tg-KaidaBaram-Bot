package repository

import (
	"github.com/jmoiron/sqlx"
	"tg-bot/internal/models"
)

const (
	users  = "users"
	events = "events"
	stats  = "stats"
)

type Auth interface {
	Create(user models.User) (int64, error)
	GetUserById(id int64) (models.User, error)
}
type Stats interface {
	Save(stat models.Statistic) error
}
type Events interface {
	Create(event models.Event, id int64) (int64, error)
	GetEvents() ([]models.Event, error)
	GetMyEvents(telegramID int64) ([]models.Event, error)
	DeleteEvent(eventID, telegramID int64) error
	SearchEvents(query string) ([]models.Event, error)
	SearchEventRandom() (models.Event, error)
	GetByID(id int64) (models.Event, error)
	RequestJoin(eventID, telegramID int64) error
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
