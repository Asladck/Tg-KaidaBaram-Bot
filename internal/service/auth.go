package service

import (
	"encoding/json"
	"tg-bot/internal/adapters/rabbitmq"
	"tg-bot/internal/models"
	"tg-bot/internal/repository"
)

type AuthService struct {
	repo   repository.Auth
	broker *rabbitmq.RabbitMQ
}

func NewAuthService(repo repository.Auth, rmq *rabbitmq.RabbitMQ) *AuthService {
	return &AuthService{repo: repo, broker: rmq}
}
func (s *AuthService) Create(user models.User) (int64, error) {
	createdUser, err := s.repo.Create(user)
	if err != nil {
		return 0, err
	}

	_, _ = s.broker.DeclareQueue("user.events")

	event := map[string]interface{}{
		"event": "user_created",
		"user":  createdUser,
	}
	body, _ := json.Marshal(event)

	// Публикуем в очередь user.events
	_ = s.broker.Publish("user.events", body)

	return createdUser, nil
}
func (s *AuthService) GetUserById(id int64) (models.User, error) {
	return s.repo.GetUserById(id)
}
