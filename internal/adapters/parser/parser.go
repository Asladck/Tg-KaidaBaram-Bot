package parser

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"tg-bot/internal/models"
)

var httpGet = http.Get

func ParseTicketonEvents() ([]models.Event, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://ticketon.kz/", nil)
	if err != nil {
		return nil, err
	}

	// Добавляем заголовки как у реального браузера
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/125.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "ru,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неожиданный статус: %s", resp.Status)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var events []models.Event

	doc.Find(".poster-item").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Find(".poster-title").Text())
		date := strings.TrimSpace(s.Find(".poster-date").Text())
		img, _ := s.Find("img").Attr("src")
		link, _ := s.Find("a").Attr("href")

		parsedDate, _ := time.Parse("02.01.2006", date)
		events = append(events, models.Event{
			Title:    title,
			Date:     parsedDate,
			ImageURL: img,
			URL:      "https://ticketon.kz" + link,
		})
	})

	log.Printf("✅ Найдено событий: %d", len(events))
	return events, nil
}
