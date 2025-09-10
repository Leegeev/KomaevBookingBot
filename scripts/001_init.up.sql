-- расширение для корректных эксклюзивных ограничений по диапазонам
CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE rooms (
  id         SERIAL PRIMARY KEY,
  name       TEXT NOT NULL,         -- 'Переговорка 1', 'Переговорка 2'
  is_active  BOOLEAN NOT NULL DEFAULT TRUE -- активна для бронирования
);

CREATE TABLE bookings (
  id          BIGSERIAL PRIMARY KEY,
  room_id     INT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
  room_name TEXT NOT NULL,  -- денормализуем для истории
  created_by  BIGINT NOT NULL,
  time_range  TSTZRANGE NOT NULL,  -- [start, end)
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
);

-- Запрещаем пересечения интервалов внутри одной переговорки.
-- NB: границы [start, end) — правая открытая, можно стыковать 10:00-11:00 и 11:00-12:00.
ALTER TABLE bookings
  ADD CONSTRAINT bookings_no_overlap
  EXCLUDE USING gist (
    room_id WITH =,
    time_range WITH &&
  );

-- По желанию: политика хранения (удалять старше 30 дней кроном) — или просто фильтровать в запросах.
