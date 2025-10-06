package repository

import (
	"github.com/jmoiron/sqlx"
	"tg-bot/internal/models"
)

type StatsPostgres struct {
	db *sqlx.DB
}

func NewStatsPostgres(db *sqlx.DB) *StatsPostgres {
	return &StatsPostgres{db: db}
}

func (r *StatsPostgres) Save(stat models.Statistic) error {
	_, err := r.db.Exec(`INSERT INTO statistics (event, data) VALUES ($1, $2)`, stat.Event, stat.Data)
	return err
}
