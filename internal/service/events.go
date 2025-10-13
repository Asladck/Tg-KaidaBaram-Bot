package service

import (
	"context"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/sirupsen/logrus"
	"tg-bot/internal/adapters/rabbitmq"
	"tg-bot/internal/models"
	"tg-bot/internal/repository"
)

type EventService struct {
	repo    repository.Events
	repAuth repository.Auth
	bot     *telego.Bot
	broker  *rabbitmq.RabbitMQ
}

func NewEventService(repo repository.Events, rmq *rabbitmq.RabbitMQ) *EventService {
	return &EventService{repo: repo, broker: rmq}
}

func (s *EventService) Create(event models.Event, telegramID int64) (int64, error) {
	id, err := s.repo.Create(event, telegramID)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func (s *EventService) GetEvents() ([]models.Event, error) {
	events, err := s.repo.GetEvents()
	if err != nil {
		logrus.Infof("Error getting events: %s", err)
		return nil, err
	}

	return events, nil
}

func (s *EventService) GetMyEvents(telegramID int64) ([]models.Event, error) {
	events, err := s.repo.GetMyEvents(telegramID)
	if err != nil {
		logrus.Infof("Error getting events: %s", err)
		return nil, err
	}
	return events, err
}

func (s *EventService) DeleteEvent(eventID, telegramID int64) error {
	err := s.repo.DeleteEvent(eventID, telegramID)
	if err != nil {
		logrus.Infof("Error deleting event: %s", err)
		return err
	}
	return nil
}

func (s *EventService) SearchEvents(query string) ([]models.Event, error) {
	events, err := s.repo.SearchEvents(query)
	if err != nil {
		logrus.Infof("Error searching events: %s", err)
		return nil, err
	}
	return events, nil
}

func (s *EventService) SearchEventRandom() (models.Event, error) {
	event, err := s.repo.SearchEventRandom()
	if err != nil {
		logrus.Infof("Error searching random event: %s", err)
		return models.Event{}, err
	}
	return event, nil
}
func (s *EventService) RequestJoin(eventID, userTgID int64) error {
	// Проверяем и сохраняем заявку
	if err := s.repo.RequestJoin(eventID, userTgID); err != nil {
		return err
	}

	// Получаем данные события
	event, err := s.repo.GetByID(eventID)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Получаем данные участника
	user, err := s.repAuth.GetUserById(userTgID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Получаем создателя события
	creator, err := s.repAuth.GetUserById(event.CreatorTgID)
	if err != nil {
		return fmt.Errorf("failed to get event creator: %w", err)
	}

	// Формируем сообщение
	text := fmt.Sprintf(
		"🆕 Новый запрос на участие!\n\nСобытие: *%s*\nОт пользователя: @%s\n\nПринять или отклонить?",
		event.Title, user.Username,
	)

	buttons := &telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{
			{
				{Text: "✅ Принять", CallbackData: fmt.Sprintf("approve_%d_%d", eventID, userTgID)},
				{Text: "❌ Отклонить", CallbackData: fmt.Sprintf("reject_%d_%d", eventID, userTgID)},
			},
		},
	}

	// Отправляем владельцу события уведомление
	if s.bot != nil {
		_, err = s.bot.SendMessage(
			context.Background(),
			&telego.SendMessageParams{
				ChatID:      telego.ChatID{ID: creator.ChatID},
				Text:        text,
				ParseMode:   "Markdown",
				ReplyMarkup: buttons,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to send Telegram message: %w", err)
		}
	}

	// Если используется брокер — дублируем уведомление

	return nil
}
