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
		INSERT INTO users (username, chat_id)
		VALUES ($1, $2)
		ON CONFLICT (chat_id) DO UPDATE 
		    SET username = EXCLUDED.username
		RETURNING id;
	`
	err := r.db.QueryRow(query, user.Username, user.ChatID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *AuthPostgres) GetUserById(chatID int64) (models.User, error) {
	var user models.User
	query := `SELECT id, username, chat_id 
			  FROM users 
			  WHERE chat_id = $1`
	err := r.db.Get(&user, query, chatID)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}
