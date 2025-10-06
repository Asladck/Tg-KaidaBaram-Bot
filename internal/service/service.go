package service

import (
	"tg-bot/internal/adapters/rabbitmq"
	"tg-bot/internal/models"
	"tg-bot/internal/repository"
)

type Auth interface {
	Create(user models.User) (int64, error)
	GetUserById(id int64) (models.User, error)
}
type Events interface {
	CheckAndUpdateEvents() error
	Recent(limit int) ([]models.Event, error)
}
type Stats interface {
	HandleEvent(body []byte) error
}
type Service struct {
	Auth
	Stats
	Events
}

func NewService(rep *repository.Repository, rmq *rabbitmq.RabbitMQ) *Service {
	return &Service{
		Auth:   NewAuthService(rep.Auth, rmq),
		Stats:  NewStatsService(rep.Stats),
		Events: NewEventService(rep.Events, rmq),
	}
}
