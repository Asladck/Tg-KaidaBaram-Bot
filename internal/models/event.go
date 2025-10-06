package models

import "time"

type Event struct {
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	Category  string    `db:"category"`
	Date      time.Time `db:"date"`
	Location  string    `db:"location"`
	URL       string    `db:"url"`        // ссылка на страницу ивента
	ImageURL  string    `db:"image_url"`  // ссылка на обложку (картинку)
	CreatedAt time.Time `db:"created_at"` // когда добавлен в БД
	UpdatedAt time.Time `db:"updated_at"` // когда последний раз обновлялся

}
