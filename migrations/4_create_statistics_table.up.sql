CREATE TABLE IF NOT EXISTS statistics (
    id SERIAL PRIMARY KEY,
    event TEXT NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
    );