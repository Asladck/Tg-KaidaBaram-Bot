# 🤖 Telegram Event Bot (Go + PostgreSQL)

**Telegram Event Bot** — это бот для управления событиями (ивентами) прямо в Telegram.  
Пользователи могут создавать, просматривать и отправлять запросы на участие в событиях.  
Проект всё ещё находится в активной разработке 🚧

## ⚙️ Стек технологий
- **Go** — основной язык разработки  
- **Telego** — Telegram Bot API библиотека  
- **PostgreSQL** — база данных  
- **Docker** — контейнеризация

## 🧩 Основные возможности
- Создание и редактирование событий  
- Просмотр списка ивентов  
- Отправка заявок на участие  
- Уведомления для создателей событий

## 🗃️ Пример таблицы `events`
```sql
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
    creator_tg_id BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'draft',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```
## 🚧 Статус

Проект в разработке, некоторые функции пока не реализованы полностью.
💡 Автор: Айбар Тлекбай
📅 2025
