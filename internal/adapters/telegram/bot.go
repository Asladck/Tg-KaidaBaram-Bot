package telegram

import (
	"context"
	"github.com/mymmrac/telego"
	"github.com/sirupsen/logrus"
)

type BotAdapter struct {
	Tg *telego.Bot
}

func NewBot(token string) (*BotAdapter, error) {
	b, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
	if err != nil {
		logrus.Errorf("telegram NewBot error: %v", err)
		return nil, err
	}
	return &BotAdapter{Tg: b}, nil
}

// Optional: helper to stop long polling if needed
func (b *BotAdapter) Close() {
	// telego specific cleanup if exists
	b.Tg.Close(context.Background()) // if telego has Close, otherwise ignore
}
