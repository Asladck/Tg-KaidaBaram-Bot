package handler

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"log"
	"strconv"
	"strings"
	"sync"
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
	mu       sync.RWMutex
}

type userState struct {
	step   string
	event  models.Event
	chatID int64
	events []models.Event
	index  int
}

func NewHandlers(bot *telego.Bot, s *service.Service) *Handlers {
	return &Handlers{
		Bot:      bot,
		Services: s,
		states:   make(map[int64]*userState),
	}
}

func (h *Handlers) Run(ctx context.Context) {
	updates, err := h.Bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		logrus.Errorf("handlers: failed to start updates via long polling: %v", err)
		return
	}
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
	if update.CallbackQuery != nil {
		callback := update.CallbackQuery.Data
		// используем chatID как идентификатор пользователя (chat id)
		chatID := update.CallbackQuery.From.ID

		if strings.HasPrefix(callback, "join_") {
			idStr := strings.TrimPrefix(callback, "join_")
			eventID, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				h.Send(chatID, "Неверный ID события")
				return
			}
			// Получаем событие и автора
			event, err := h.Services.Events.GetByID(eventID)
			if err != nil {
				h.Send(chatID, "Ошибка при получении события")
				return
			}
			// Получаем автора (используем creator_chat_id)
			author, err := h.Services.GetUserById(event.CreatorTgID)
			if err != nil {
				logrus.Infof("Error getting event author: %s", err)
				h.Send(chatID, "Ошибка при получении автора события")
				return
			}
			// Получаем пользователя, который хочет присоединиться (используем chatID)
			user, err := h.Services.GetUserById(chatID)
			if err != nil {
				h.Send(chatID, "Ошибка при получении пользователя")
				return
			}
			// Отправляем автору сообщение
			msg := fmt.Sprintf("Пользователь @%s (id=%d) хочет присоединиться к вашему событию: %s", user.Username, user.ChatID, event.Title)
			h.Send(author.ChatID, msg)
			h.Send(chatID, "Запрос на участие отправлен автору события!")
			return
		}
		if strings.HasPrefix(callback, "next_") {
			idStr := strings.TrimPrefix(callback, "next_")
			eventID, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				h.Send(chatID, "Неверный ID события")
				return
			}
			h.handleNextCommand(chatID, eventID)
			return
		}
		return
	}

	if update.Message == nil {
		return
	}
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	if strings.HasPrefix(text, "/apply_") {
		idStr := strings.TrimPrefix(text, "/apply_")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			h.Send(chatID, "Неверный ID события")
			return
		}
		h.handleApplyCommand(chatID, id)
		return
	}
	if strings.HasPrefix(text, "/next_") {
		idStr := strings.TrimPrefix(text, "/next_")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			h.Send(chatID, "Неверный ID события")
			return
		}
		h.handleNextCommand(chatID, id)
		return
	}

	switch text {
	case "/start":
		h.handleStart(chatID, update.Message.From.Username)
	case "/create":
		h.handleCreateCommand(chatID)
	case "/events":
		h.handleEventsCommand(chatID)
	case "/my_events":
		h.handleMyEventsCommand(chatID)
	case "/search":
		h.handleSearchCommand(chatID)
	case "/random":
		h.handleRandomCommand(chatID)

	default:
		h.handleUserState(chatID, text)
	}
}
func (h *Handlers) handleSearchCommand(chatID int64) {
	user, err := h.Services.GetUserById(chatID)
	if err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}

	h.Send(chatID, "🔍 Введите ключевое слово для поиска в названиях событий:")
	h.mu.Lock()
	h.states[user.ChatID] = &userState{step: "search_keyword", chatID: user.ChatID} // Сохраняем состояние по chatID
	h.mu.Unlock()
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
	months := []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
	msg := fmt.Sprintf("Случайное событие:\nID: %d\nНазвание: %s\nКатегория: %s\nДата: %d %s\nМесто: %s\nСсылка: %s\n",
		event.ID, event.Title, event.Category, event.Date.Day(), months[event.Date.Month()-1], event.Location, event.URL)
	h.Send(chatID, msg)
}

func (h *Handlers) handleStart(chatID int64, username string) {
	user := models.User{
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

func (h *Handlers) handleCreateCommand(chatID int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}
	h.mu.Lock()
	h.states[chatID] = &userState{step: "title", chatID: chatID}
	h.mu.Unlock()
	h.Send(chatID, "🎬 Введите название мероприятия:")
}
func (h *Handlers) handleMyEventsCommand(chatID int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}
	h.sendMyEventsList(chatID)
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
		logrus.Infof("Error searching events: %s", err)
		h.Send(chatID, "Ошибка при поиске событий 😢")
		return
	}
	user, err := h.Services.GetUserById(chatID)
	if err != nil {
		logrus.Infof("Error getting user: %s", err)
		h.Send(chatID, "Ошибка при поиске событий 😢")
		return
	}

	if len(events) == 0 {
		h.Send(chatID, fmt.Sprintf("❌ Не найдено событий по запросу: %s", keyword))
		return
	}

	h.mu.Lock()
	h.states[user.ChatID] = &userState{
		step:   "browse_events",
		chatID: user.ChatID,
		events: events,
		index:  0,
	}
	h.mu.Unlock()

	// Показываем первое событие и устанавливаем индекс
	h.sendEventByIndex(chatID, 0)
}
func (h *Handlers) sendEventByIndex(chatID int64, index int) {
	h.mu.RLock()
	state, ok := h.states[chatID]
	h.mu.RUnlock()
	if !ok || state == nil || index < 0 || index >= len(state.events) {
		h.Send(chatID, "События закончились 🔚")
		return
	}

	event := state.events[index]
	msg := fmt.Sprintf("📌 Событие %d из %d:\n\nНазвание: %s\nКатегория: %s\n📅 Дата: %s\n📍 Место: %s\n🔗 Ссылка: %s",
		index+1, len(state.events),
		event.Title, event.Category,
		event.Date.Format("02.01.2006"), event.Location, event.URL)

	// Inline-кнопки
	buttons := [][]telego.InlineKeyboardButton{
		{
			telego.InlineKeyboardButton{
				Text:         "Запросить участие",
				CallbackData: fmt.Sprintf("join_%d", event.ID),
			},
			telego.InlineKeyboardButton{
				Text:         "Следующий",
				CallbackData: fmt.Sprintf("next_%d", event.ID),
			},
		},
	}
	keyboard := telego.InlineKeyboardMarkup{InlineKeyboard: buttons}

	// Обновляем текущий индекс безопасно
	h.mu.Lock()
	if s, ok := h.states[chatID]; ok && s != nil {
		s.index = index
		h.states[chatID] = s
	}
	h.mu.Unlock()

	_, err := h.Bot.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID:      telego.ChatID{ID: chatID},
		Text:        msg,
		ReplyMarkup: &keyboard,
	})
	if err != nil {
		logrus.Errorf("Ошибка отправки сообщения с кнопками: %v", err)
	}
}

func (h *Handlers) handleApplyCommand(chatID, eventID int64) {
	if _, err := h.Services.GetUserById(chatID); err != nil {
		h.Send(chatID, "Привет, Гость! Тебе нужно зарегистрироваться! \n /start <- Нажми")
		return
	}
	err := h.Services.RequestJoin(eventID, chatID)
	if err != nil {
		logrus.Infof("Error applying to event: %s", err)
		h.Send(chatID, "Ошибка при отправке заявки 😢")
		return
	}
	h.Send(chatID, fmt.Sprintf("✅ Ваша заявка на участие в событии ID %d отправлена!", eventID))
}
func (h *Handlers) handleNextCommand(chatID, eventID int64) {
	h.mu.RLock()
	state, ok := h.states[chatID]
	h.mu.RUnlock()
	if !ok || state == nil || len(state.events) == 0 {
		h.Send(chatID, "Нет активного поиска. Введите /search чтобы начать снова 🔍")
		return
	}

	// Попробуем найти текущий индекс по eventID (кнопка /next_<id>)
	currentIndex := state.index
	found := false
	for i, ev := range state.events {
		if ev.ID == eventID {
			currentIndex = i
			found = true
			break
		}
	}
	// если не нашли по id, используем сохранённый индекс

	nextIndex := currentIndex + 1
	if nextIndex >= len(state.events) {
		h.Send(chatID, "Больше событий нет 😢")
		h.mu.Lock()
		delete(h.states, chatID)
		h.mu.Unlock()
		return
	}

	// Обновляем индекс и показываем следующее событие
	h.mu.Lock()
	if s, ok := h.states[chatID]; ok && s != nil {
		s.index = nextIndex
		h.states[chatID] = s
	}
	h.mu.Unlock()

	h.sendEventByIndex(chatID, nextIndex)

	// Если не найдено по id и не было активности — ничего дополнительного не делаем
	_ = found
}
func (h *Handlers) handleUserState(chatID int64, text string) {
	h.mu.RLock()
	state, ok := h.states[chatID]
	h.mu.RUnlock()
	if !ok || state == nil {
		return
	}

	// Блокируем на запись, потому что будем менять состояние
	h.mu.Lock()
	defer h.mu.Unlock()

	// Нужно снова проверить, т.к. состояние могло измениться до блокировки
	state, ok = h.states[chatID]
	if !ok || state == nil {
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
		// Сохраняем текущий ключевый запрос временно, но обработка поиска отправит новое состояние
		// Разблокируем и вызовем обработчик поиска вне lock, чтобы избежать двойной блокировки по h.mu
		keyword := text
		delete(h.states, chatID)
		h.mu.Unlock()
		h.handleSearchKeyword(chatID, keyword)
		h.mu.Lock()
	case "choose_action":
		// Здесь можно обработать выбор действия, например, отправку заявки или просмотр следующего события
		h.Send(chatID, "Выберите действие: 1. Отправить заявку (/apply_<id>) 2. Следующий ивент (/next_<id>)")
		// После обработки действия можно удалить состояние пользователя
		delete(h.states, chatID)
		return
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
		evID, err := h.Services.Events.Create(state.event, chatID)
		if err != nil {
			logrus.Infof("Error: %s", err.Error())
			h.Send(chatID, "Ошибка при создании события")
			delete(h.states, chatID)
			return
		}
		h.Send(chatID, fmt.Sprintf("✅ Создано! ID: %d, %s — %s", evID, state.event.Title, state.event.Location))
		delete(h.states, chatID)
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
	months := []string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
	for i, event := range events {
		msg := fmt.Sprintf("Событие %d:\nID: %d\nНазвание: %s\nКатегория: %s\nДата:%d %s\nМесто: %s\nСсылка: %s\n",
			i+1, event.ID, event.Title, event.Category, event.Date.Day(), months[event.Date.Month()-1], event.Location, event.URL)
		h.Send(chatID, msg)
	}
}
func (h *Handlers) sendMyEventsList(chatID int64) {
	events, err := h.Services.Events.GetMyEvents(chatID)
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
			i+1, event.ID, event.Title, event.Category, event.Date.Format("02.01.2006"), event.Location, event.URL)
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
