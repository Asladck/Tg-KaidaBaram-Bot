package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"tg-bot/internal/models"
)

type EventPostgres struct {
	db *sqlx.DB
}

func NewEventPostgres(db *sqlx.DB) *EventPostgres {
	return &EventPostgres{db: db}
}

func (r *EventPostgres) Create(event models.Event, telegramID int64) (int64, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 1️⃣ Находим ID пользователя по telegram_id
	var userID int64
	queryUser := `SELECT id FROM users WHERE telegram_id = $1`
	err = tx.Get(&userID, queryUser, telegramID)
	if err != nil {
		return 0, fmt.Errorf("user with telegram_id=%d not found: %w", telegramID, err)
	}

	// 2️⃣ Создаём событие
	var eventID int64
	queryEvent := `
		INSERT INTO events (title, category, date, location, description, url, image_url, creator_id,creator_telegram_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'draft', NOW(), NOW())
		RETURNING id
	`
	err = tx.QueryRow(queryEvent,
		event.Title,
		event.Category,
		event.Date,
		event.Location,
		event.Description,
		event.URL,
		event.ImageURL,
		userID,
		telegramID,
	).Scan(&eventID)
	if err != nil {
		return 0, err
	}

	return eventID, nil
}

func (r *EventPostgres) GetEvents() ([]models.Event, error) {
	var eventsList []models.Event
	query := fmt.Sprintf(`SELECT id, title, category, date, location, description ,url, image_url, creator_id, created_at, updated_at, status FROM %s WHERE date >= NOW() ORDER BY date`, events)
	err := r.db.Select(&eventsList, query)
	if err != nil {
		return nil, err
	}
	return eventsList, nil
}

func (r *EventPostgres) GetMyEvents(telegramID int64) ([]models.Event, error) {
	var eventsList []models.Event

	query := `
		SELECT e.id, e.title, e.category, e.date, e.location, e.description, e.url, e.image_url, e.creator_id, e.created_at, e.updated_at, e.status
		FROM events e
		JOIN users u ON e.creator_id = u.id
		WHERE u.telegram_id = $1
		ORDER BY e.date
	`
	err := r.db.Select(&eventsList, query, telegramID)
	if err != nil {
		return nil, err
	}
	return eventsList, nil
}
func (r *EventPostgres) DeleteEvent(eventID, telegramID int64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	// 1️⃣ Находим ID пользователя по telegram_id
	var userID int64
	queryUser := `SELECT id FROM users WHERE telegram_id = $1`
	err = tx.Get(&userID, queryUser, telegramID)
	if err != nil {
		return fmt.Errorf("user with telegram_id=%d not found: %w", telegramID, err)
	}

	// 2️⃣ Удаляем событие, если оно принадлежит пользователю
	queryDelete := `DELETE FROM events WHERE id = $1 AND creator_id = $2`
	result, err := tx.Exec(queryDelete, eventID, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("event with id=%d not found or does not belong to user with telegram_id=%d", eventID, telegramID)
	}

	return nil
}

func (r *EventPostgres) SearchEvents(query string) ([]models.Event, error) {
	var eventsList []models.Event
	searchQuery := fmt.Sprintf(`SELECT id, title, category, date, location, description ,url, image_url, creator_id, created_at, updated_at, status 
		FROM %s 
		WHERE (title ILIKE '%%' || $1 || '%%' OR description ILIKE '%%' || $1 || '%%') AND date >= NOW() 
		ORDER BY date`, events)
	err := r.db.Select(&eventsList, searchQuery, query)
	if err != nil {
		return nil, err
	}
	return eventsList, nil
}

func (r *EventPostgres) SearchEventRandom() (models.Event, error) {
	var event models.Event
	query := fmt.Sprintf(`SELECT id, title, category, date, location, description ,url, image_url, creator_id, created_at, updated_at, status 
		FROM %s 
		WHERE date >= NOW() 
		ORDER BY RANDOM() 
		LIMIT 1`, events)
	err := r.db.Get(&event, query)
	if err != nil {
		return models.Event{}, err
	}
	return event, nil
}

func (r *EventPostgres) GetByID(id int64) (models.Event, error) {
	var event models.Event
	query := fmt.Sprintf(`SELECT id, title, category, date, location, description ,url, image_url, creator_id, created_at, updated_at, status 
		FROM %s 
		WHERE id = $1`, events)
	err := r.db.Get(&event, query, id)
	if err != nil {
		return models.Event{}, err
	}
	return event, nil
}
func (r *EventPostgres) RequestJoin(eventID, userTgID int64) error {
	// Проверяем, существует ли событие
	var exists bool
	queryEvent := `SELECT EXISTS(SELECT 1 FROM events WHERE id = $1 AND date >= NOW())`
	err := r.db.Get(&exists, queryEvent, eventID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("event with id=%d does not exist or has already occurred", eventID)
	}

	// Проверяем, существует ли пользователь
	var userID int64
	queryUser := `SELECT id FROM users WHERE telegram_id = $1`
	err = r.db.Get(&userID, queryUser, userTgID)
	if err != nil {
		return fmt.Errorf("user with telegram_id=%d not found: %w", userTgID, err)
	}

	// Проверяем, не является ли пользователь создателем события
	var creatorID int64
	queryCreator := `SELECT creator_id FROM events WHERE id = $1`
	err = r.db.Get(&creatorID, queryCreator, eventID)
	if err != nil {
		return err
	}
	if creatorID == userID {
		return fmt.Errorf("user with telegram_id=%d is the creator of the event and cannot join it", userTgID)
	}

	// Проверяем, не отправлял ли пользователь уже заявку на это событие
	var requestExists bool
	queryRequest := `SELECT EXISTS(SELECT 1 FROM event_requests WHERE event_id = $1 AND user_id = $2)`
	err = r.db.Get(&requestExists, queryRequest, eventID, userID)
	if err != nil {
		return err
	}
	if requestExists {
		return fmt.Errorf("user with telegram_id=%d has already requested to join event with id=%d", userTgID, eventID)
	}

	// Сохраняем заявку на участие
	queryInsert := `INSERT INTO event_requests (event_id, user_id, requested_at) VALUES ($1, $2, NOW())`
	_, err = r.db.Exec(queryInsert, eventID, userID)
	if err != nil {
		return err
	}

	return nil
}
