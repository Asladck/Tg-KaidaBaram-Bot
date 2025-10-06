package repository

import (
	"github.com/jmoiron/sqlx"
	"tg-bot/internal/models"
)

type AuthPostgres struct {
	db *sqlx.DB
}

func NewAuthPostgres(db *sqlx.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

func (r *AuthPostgres) Create(user models.User) (int64, error) {
	var id int64
	query := `
		INSERT INTO users (telegram_id, username, chat_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (telegram_id) DO UPDATE 
		    SET username = EXCLUDED.username,
		        chat_id = EXCLUDED.chat_id
		RETURNING id;
	`
	err := r.db.QueryRow(query, user.TgID, user.Username, user.ChatID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthPostgres) GetUserById(id int64) (models.User, error) {
	var user models.User
	query := `SELECT id, telegram_id, username, chat_id 
			  FROM users 
			  WHERE id = $1`
	err := r.db.Get(&user, query, id)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
