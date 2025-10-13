CREATE TABLE IF NOT EXISTS event_participants (
    id SERIAL PRIMARY KEY,
    event_id INT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected
    requested_at TIMESTAMP DEFAULT NOW(),
    confirmed_at TIMESTAMP,
    UNIQUE (event_id, user_id)
    );
