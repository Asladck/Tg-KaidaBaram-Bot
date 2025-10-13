package handler

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"time"

	"github.com/mymmrac/telego"
	"tg-bot/internal/models"
	"tg-bot/internal/service"
)

type Messenger interface {
	SendMessage(chatID int64, text string)
}

type Handlers struct {
	Bot      *telego.Bot
	Services *service.Service
	states   map[int64]*userState
}

type userState struct {
	step   string
	event  models.Event
	userID int64
}

func NewHandlers(bot *telego.Bot, s *service.Service) *Handlers {
	return &Handlers{
		Bot:      bot,
		Services: s,
		states:   make(map[int64]*userState),
	}
}

func (h *Handlers) Run(ctx context.Context) {
	updates, _ := h.Bot.UpdatesViaLongPolling(ctx, nil)
	for {
		select {
		case <-ctx.Done():
			log.Println("handlers: ctx canceled, stopping handlers.Run")
			return
		case update, ok := <-updates:
			if !ok {
				log.Println("handlers: updates channel closed")
				return
			}
			go h.handleUpdate(update)
		}
	}
}

func (h *Handlers) handleUpdate(update telego.Update) {
	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	tgID := update.Message.From.ID
	text := update.Message.Text

	switch text {
	case "/start":
		h.handleStart(chatID, tgID, update.Message.From.Username)
	case "/create":
		h.handleCreateCommand(chatID, tgID)
	case "/events":
		h.handleEventsCommand(chatID)
	case "/my_events":
		h.handleMyEventsCommand(chatID, tgID)
	case "/search":
		h.handleSearchCommand(chatID)
	case "/random":
		h.handleRandomCommand(chatID)

	default:
		h.handleUserState(chatID, tgID, text)
	}
}
func (h *Handlers) handleSearchCommand(chatID int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}

	h.Send(chatID, "🔍 Введите ключевое слово для поиска в названиях событий:")
	h.states[chatID] = &userState{step: "search_keyword"} // Сохраняем состояние
}

func (h *Handlers) handleRandomCommand(chatID int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}
	h.sendRandomEvent(chatID)
}
func (h *Handlers) sendRandomEvent(chatID int64) {
	event, err := h.Services.SearchEventRandom()
	if err != nil {
		logrus.Infof("Error getting random event: %s", err)
		h.Send(chatID, "Ошибка при получении случайного события")
		return
	}
	if event.ID == 0 {
		h.Send(chatID, "Событий нет")
		return
	}
	msg := fmt.Sprintf("Случайное событие:\nID: %d\nНазвание: %s\nКатегория: %s\nДата: %s\nМесто: %s\nСсылка: %s\n",
		event.ID, event.Title, event.Category, event.Date.Format("января февраля марта апреля мая июня июля августа сентября октября ноября декабря")[event.Date.Month()*8-8:event.Date.Month()*8], event.Location, event.URL)
	h.Send(chatID, msg)
}

func (h *Handlers) handleStart(chatID, tgID int64, username string) {
	user := models.User{
		TgID:     tgID,
		Username: username,
		ChatID:   chatID,
	}
	u, err := h.Services.Auth.Create(user)
	if err != nil {
		h.Send(chatID, "Ошибка при регистрации")
		return
	}
	h.Send(chatID, fmt.Sprintf("Привет, %s! Ты зарегистрирован (id=%d)", user.Username, u))
}

func (h *Handlers) handleCreateCommand(chatID, tgID int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}
	h.states[tgID] = &userState{step: "title", userID: tgID}
	h.Send(chatID, "🎬 Введите название мероприятия:")
}
func (h *Handlers) handleMyEventsCommand(chatID, telegramId int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}
	h.sendMyEventsList(chatID, telegramId)
}
func (h *Handlers) handleEventsCommand(chatID int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}
	h.sendEventsList(chatID)
}
func (h *Handlers) handleSearchKeyword(chatID int64, keyword string) {
	events, err := h.Services.SearchEvents(keyword)
	if err != nil {
		h.Send(chatID, "Ошибка при поиске событий 😢")
		return
	}

	if len(events) == 0 {
		h.Send(chatID, fmt.Sprintf("❌ Не найдено событий по запросу: %s", keyword))
		return
	}

	h.Send(chatID, fmt.Sprintf("🔎 Найдено %d событий по запросу '%s':", len(events), keyword))
	for i, event := range events {
		msg := fmt.Sprintf("Событие %d:\n📌 Название: %s\nКатегория: %s\n📅 Дата: %s\n📍 Место: %s\n🔗 Ссылка: %s\n",
			i+1, event.Title, event.Category, event.Date.Format("02.01.2006"), event.Location, event.URL)
		h.Send(chatID, msg)
	}
}

func (h *Handlers) handleUserState(chatID, tgID int64, text string) {
	state, ok := h.states[tgID]
	if !ok {
		return
	}
	switch state.step {
	case "title":
		state.event.Title = text
		state.step = "category"
		h.Send(chatID, "🗂 Введите категорию:")
	case "category":
		state.event.Category = text
		state.step = "description"
		h.Send(chatID, "📝 Введите описание:")
	case "search_keyword":
		h.handleSearchKeyword(chatID, text)
		delete(h.states, tgID)
	case "description":
		state.event.Description = text
		state.step = "date"
		h.Send(chatID, "📅 Введите дату (YYYY-MM-DD):")
	case "date":
		parsed, err := time.Parse("2006-01-02", text)
		if err != nil {
			h.Send(chatID, "Неверный формат даты. Попробуйте YYYY-MM-DD")
			return
		}
		state.event.Date = parsed
		state.step = "location"
		h.Send(chatID, "📍 Укажите место:")
	case "location":
		state.event.Location = text
		state.step = "url"
		h.Send(chatID, "🔗 Вставьте ссылку на событие (необязательно):")
	case "url":
		state.event.URL = text
		state.event.CreatedAt = time.Now()
		state.event.UpdatedAt = time.Now()
		evID, err := h.Services.Events.Create(state.event, tgID)
		if err != nil {
			logrus.Infof("Error: %s", err.Error())
			h.Send(chatID, "Ошибка при создании события")
			delete(h.states, tgID)
			return
		}
		h.Send(chatID, fmt.Sprintf("✅ Создано! ID: %d, %s — %s", evID, state.event.Title, state.event.Location))
		delete(h.states, tgID)
	}
}

func (h *Handlers) sendEventsList(chatID int64) {
	events, err := h.Services.Events.GetEvents()
	if err != nil {
		logrus.Infof("Error getting events: %s", err)
		h.Send(chatID, "Ошибка при получении событий")
		return
	}
	if len(events) == 0 {
		h.Send(chatID, "Событий нет")
		return
	}
	for i, event := range events {
		msg := fmt.Sprintf("Событие %d:\nID: %d\nНазвание: %s\nКатегория: %s\nДата: %s\nМесто: %s\nСсылка: %s\n",
			i+1, event.ID, event.Title, event.Category, event.Date.Format("января февраля марта апреля мая июня июля августа сентября октября ноября декабря")[event.Date.Month()*8-8:event.Date.Month()*8], event.Location, event.URL)
		h.Send(chatID, msg)
	}
}
func (h *Handlers) sendMyEventsList(chatID, telegramId int64) {
	events, err := h.Services.Events.GetMyEvents(telegramId)
	if err != nil {
		logrus.Infof("Error getting events: %s", err)
		h.Send(chatID, "Ошибка при получении событий")
		return
	}
	if len(events) == 0 {
		h.Send(chatID, "Событий нет")
		return
	}
	h.Send(chatID, fmt.Sprintf("Ваши события (всего: %d):\n", len(events)))
	for i, event := range events {
		msg := fmt.Sprintf("Событие %d:\nID: %d\nНазвание: %s\nКатегория: %s\nДата: %s\nМесто: %s\nСсылка: %s\n",
			i+1, event.ID, event.Title, event.Category, event.Date.Format("января февраля марта апреля мая июня июля августа сентября октября ноября декабря")[event.Date.Month()*8-8:event.Date.Month()*8], event.Location, event.URL)
		h.Send(chatID, msg)
	}
}

// Send — обертка для отправки сообщений
func (h *Handlers) Send(chatID int64, text string) {
	_, err := h.Bot.SendMessage(
		context.Background(), // 👈 добавляем контекст
		&telego.SendMessageParams{
			ChatID: telego.ChatID{ID: chatID},
			Text:   text,
		},
	)
	if err != nil {
		fmt.Println("Ошибка при отправке сообщения:", err)
	}
}
