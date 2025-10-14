CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT,
    date TIMESTAMP NOT NULL,
    location TEXT,
    description TEXT,
    url TEXT,
    image_url TEXT,
    creator_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    creator_telegram_id BIGINT NOT NULL REFERENCES users(chat_id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'draft', -- draft, published, closed
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
    );
