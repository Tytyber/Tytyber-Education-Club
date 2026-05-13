CREATE TABLE IF NOT EXISTS users (
                                     id BIGSERIAL PRIMARY KEY,
                                     email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    username VARCHAR(100),
    xp INT DEFAULT 0,
    level INT DEFAULT 1,
    streak INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

CREATE INDEX idx_users_email ON users(email);

-- Таблица Roadmap (Карьерный путь)
CREATE TABLE IF NOT EXISTS roadmaps (
                                        id BIGSERIAL PRIMARY KEY,
                                        user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL, -- Например "Go Backend Developer"
    status VARCHAR(20) DEFAULT 'active', -- active, completed
    progress_percent INT DEFAULT 0,
    ai_raw_data JSONB, -- Сюда сохраним сырой ответ от AI для гибкости
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

-- Этапы внутри Roadmap (Ступени)
CREATE TABLE IF NOT EXISTS roadmap_stages (
                                              id BIGSERIAL PRIMARY KEY,
                                              roadmap_id BIGINT REFERENCES roadmaps(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL, -- "Basic Syntax", "Concurrency"
    order_idx INT NOT NULL,
    status VARCHAR(20) DEFAULT 'locked', -- locked, unlocked, completed
    estimated_hours INT
    );