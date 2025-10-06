package service

import (
	"log"
	"tg-bot/internal/adapters/parser"
	"tg-bot/internal/models"
	"time"

	"tg-bot/internal/adapters/rabbitmq"
	"tg-bot/internal/repository"
)

type EventService struct {
	repo   repository.Events
	broker *rabbitmq.RabbitMQ
}

func NewEventService(repo repository.Events, rmq *rabbitmq.RabbitMQ) *EventService {
	return &EventService{repo: repo, broker: rmq}
}

func (s *EventService) Recent(limit int) ([]models.Event, error) {
	events, err := s.repo.GetRecentEvents(limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// CheckAndUpdateEvents парсит сайт, обновляет БД
func (s *EventService) CheckAndUpdateEvents() error {
	// 1. Получаем актуальные события с сайта
	events, err := parser.ParseTicketonEvents()
	if err != nil {
		return err
	}

	// 2. Сохраняем новые события
	for _, e := range events {
		exists, _ := s.repo.Exists(e.Title)
		if !exists {
			if err := s.repo.AddEvent(e); err != nil {
				log.Printf("Ошибка добавления %s: %v", e.Title, err)
				continue
			}
			// Отправляем событие в RabbitMQ
			err := s.broker.Publish("events.new", []byte(e.Title))
			if err != nil {
				return err
			}
			log.Printf("Добавлено новое событие: %s", e.Title)
		}
	}

	// 3. Удаляем прошедшие события
	if err := s.repo.DeleteOldEvents(time.Now()); err != nil {
		log.Printf("Ошибка удаления старых событий: %v", err)
	}

	return nil
}
