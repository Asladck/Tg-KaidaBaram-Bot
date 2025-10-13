CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT UNIQUE NOT NULL,
    username VARCHAR(255),
    chat_id BIGINT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );
