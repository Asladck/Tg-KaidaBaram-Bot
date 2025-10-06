package telegram

import (
	"context"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/sirupsen/logrus"
	"tg-bot/internal/models"
	"tg-bot/internal/service"
)

type Bot struct {
	Tg       *telego.Bot
	services *service.Service
}

func InitBot(token string, service *service.Service) (*Bot, error) {
	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		logrus.Infof("Error with integrating BOT : %s", err.Error())
		return nil, err
	}

	return &Bot{Tg: bot, services: service}, nil
}
func (b *Bot) SendMessage(chatID int64, text string) error {
	_, err := b.Tg.SendMessage(context.Background(), &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   text,
	})
	return err
}
func (b *Bot) Start() {
	updates, _ := b.Tg.UpdatesViaLongPolling(context.Background(), nil)
	for update := range updates {
		if update.Message != nil && update.Message.Text == "/start" {
			chatID := update.Message.Chat.ID
			user := models.User{
				TgID:     update.Message.From.ID,
				Username: update.Message.From.Username,
				ChatID:   chatID,
				Category: nil,
			}
			id, err := b.services.Create(user)
			if err != nil {
				logrus.Infof(err.Error())
				_ = b.SendMessage(chatID, "Ошибка при регистрации.")
			} else {
				_ = b.SendMessage(chatID, fmt.Sprintf("Привет, %s! Ты зарегистрирован (id=%d)", user.Username, id))
			}
		}
	}
}
func (b *Bot) HandleRecentEvents(chatID int64) {
	events, err := b.services.Recent(5) // последние 5 событий
	if err != nil {
		_ = b.SendMessage(chatID, "⚠️ Ошибка при получении событий.")
		return
	}

	if len(events) == 0 {
		_ = b.SendMessage(chatID, "Пока нет новых событий 🎭")
		return
	}

	for _, e := range events {
		msg := fmt.Sprintf("🎫 *%s*\n📅 %s\n🔗 [Подробнее](%s)",
			e.Title, e.Date.Format("02 Jan 2006"), e.URL)

		photo := &telego.SendPhotoParams{
			ChatID:    telego.ChatID{ID: chatID},
			Caption:   msg,
			ParseMode: "Markdown",
		}

		_, err := b.Tg.SendPhoto(context.Background(), photo)
		if err != nil {
			logrus.Warnf("Ошибка отправки фото: %v", err)
		}
	}
}

func CheckBot(bot *Bot) error {
	botUser, err := bot.Tg.GetMe(context.Background())
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	fmt.Printf("Bot user: %+v\n", botUser)
	return nil
}
