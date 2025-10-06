package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"tg-bot/internal/models"
	"tg-bot/internal/repository"
)

type StatsService struct {
	repo repository.Stats
}

func NewStatsService(repo repository.Stats) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) HandleEvent(body []byte) error {
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		logrus.Errorf("failed to unmarshal event: %s", err)
		return err
	}

	// Пример: сохраняем в БД
	stat := models.Statistic{
		Event: event["event"].(string),
		Data:  string(body),
	}
	if err := s.repo.Save(stat); err != nil {
		logrus.Errorf("failed to save statistic: %s", err)
		return err
	}

	logrus.Infof("Stat saved: %+v", stat)
	return nil
}
