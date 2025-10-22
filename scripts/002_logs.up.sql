-- ===============================================
-- 002_logs.up.sql
-- Таблицы для хранения пользователей, запросов и соглашений
-- ===============================================

-- 1️⃣ Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id          BIGINT PRIMARY KEY,            -- Telegram user_id
    fio         TEXT NOT NULL,                 -- ФИО пользователя
    created_at  TIMESTAMPTZ DEFAULT now()      -- когда был добавлен
);

-- 2️⃣ Таблица запросов
CREATE TABLE IF NOT EXISTS zaprosy (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name   TEXT NOT NULL,                 -- username из Telegram (на момент создания)
    date        DATE NOT NULL,                 -- дата запроса
    doveritel   TEXT,                          -- доверитель, если указан
    comment     TEXT,                          -- произвольный комментарий
    created_at  TIMESTAMPTZ DEFAULT now()      -- когда был создан запрос
);

-- Индекс для быстрого поиска по пользователю и дате
CREATE INDEX IF NOT EXISTS idx_zaprosy_user_created_at
    ON zaprosy (user_id, created_at DESC);

-- 3️⃣ Таблица соглашений
CREATE TABLE IF NOT EXISTS soglasheniya (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_name   TEXT NOT NULL,                 -- username из Telegram (на момент создания)
    date        DATE NOT NULL,                 -- дата соглашения
    doveritel   TEXT,                          -- доверитель
    comment     TEXT,                          -- комментарий
    created_at  TIMESTAMPTZ DEFAULT now()      -- когда было создано соглашение
);

-- Индекс для быстрого поиска по пользователю и дате
CREATE INDEX IF NOT EXISTS idx_soglasheniya_user_created_at
    ON soglasheniya (user_id, created_at DESC);