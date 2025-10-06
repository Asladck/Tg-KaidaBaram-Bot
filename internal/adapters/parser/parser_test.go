package parser

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tg-bot/internal/models"
)

const sampleHTML = `
<div class="poster-item">
  <img src="/images/event1.jpg">
  <a href="/event/123"></a>
  <div class="poster-title">Концерт Айым</div>
  <div class="poster-date">05.10.2025</div>
</div>
<div class="poster-item">
  <img src="/images/event2.jpg">
  <a href="/event/456"></a>
  <div class="poster-title">Фестиваль SDU</div>
  <div class="poster-date">10.10.2025</div>
</div>
`

func TestParseTicketonEvents(t *testing.T) {
	// Поднимаем тестовый HTTP-сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, strings.NewReader(sampleHTML))
	}))
	defer ts.Close()

	// Подменяем httpGet на локальный сервер
	oldGet := httpGet
	httpGet = func(url string) (*http.Response, error) {
		return http.Get(ts.URL)
	}
	defer func() { httpGet = oldGet }()

	// Запускаем тест
	events, err := ParseTicketonEvents()
	if err != nil {
		t.Fatalf("ParseTicketonEvents вернула ошибку: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("ожидалось 2 события, получено %d", len(events))
	}

	expected := models.Event{
		Title:    "Концерт Айым",
		Date:     time.Date(2025, 10, 5, 0, 0, 0, 0, time.UTC),
		ImageURL: "/images/event1.jpg",
		URL:      "https://ticketon.kz/event/123",
	}

	got := events[0]
	if got.Title != expected.Title {
		t.Errorf("Title: ожидался %q, получен %q", expected.Title, got.Title)
	}
	if got.ImageURL != expected.ImageURL {
		t.Errorf("ImageURL: ожидался %q, получен %q", expected.ImageURL, got.ImageURL)
	}
	if got.URL != expected.URL {
		t.Errorf("URL: ожидался %q, получен %q", expected.URL, got.URL)
	}
}
