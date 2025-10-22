package repository

const qSchedule = `
	SELECT
		r.name,
		lower(b.time_range AT TIME ZONE 'Europe/Moscow') AS start_local,
		upper(b.time_range AT TIME ZONE 'Europe/Moscow') AS end_local
	FROM bookings b
	JOIN rooms r ON r.id = b.room_id
	WHERE b.time_range && tstzrange(
		date_trunc('day', now() AT TIME ZONE 'Europe/Moscow') AT TIME ZONE 'Europe/Moscow',
		(date_trunc('day', now() AT TIME ZONE 'Europe/Moscow') + interval '1 day') AT TIME ZONE 'Europe/Moscow',
		'[)'
	)
	ORDER BY r.name, start_local;
`

// BOOKING REPOSITORY QUERIES

const qInsertBooking = `
INSERT INTO bookings (room_id, room_name, user_id, user_name, time_range)
VALUES ($1, $2, $3, $4, tstzrange($5, $6, '[)'))
RETURNING id;
`

const qDeleteByID = `
DELETE FROM bookings
WHERE id = $1;
`

const qSelectByID = `
SELECT
  id,
  room_id,
  room_name,
  user_id,
  user_name,
  lower(time_range) AS start_utc,
  upper(time_range) AS end_utc,
  created_at
FROM bookings
WHERE id = $1;
`

const qListByRoomAndInterval = `
SELECT
  id,
  room_id,
  room_name,
  user_id,
  user_name,
  lower(time_range) AS start_utc,
  upper(time_range) AS end_utc,
  created_at
FROM bookings
WHERE room_id = $1
  AND time_range && tstzrange($2, $3, '[)')
ORDER BY lower(time_range) ASC, id ASC;
`

// future = все, у кого верхняя граница в будущем относительно fromUTC
const qListByUser = `
SELECT
  id,
  room_id,
  room_name,
  user_id,
  user_name,
  lower(time_range) AS start_utc,
  upper(time_range) AS end_utc,
  created_at
FROM bookings
WHERE user_id = $1
  AND upper(time_range) > $2
ORDER BY lower(time_range) ASC, id ASC;
`

const qAnyOverlap = `
SELECT EXISTS (
  SELECT 1
  FROM bookings
  WHERE room_id = $1
    AND time_range && tstzrange($2, $3, '[)')
) AS overlap;
`

const qDeleteEndedBefore = `
DELETE FROM bookings
WHERE upper(time_range) < $1;
`

// ROOM REPOSITORY QUERIES

const qInsertRoom = `
INSERT INTO rooms (name, is_active)
VALUES ($1, $2)
RETURNING id;
`

// "Удаление" = деактивация (идемпотентно: активную делаем неактивной)
const qDeactivateRoom = `
UPDATE rooms
SET is_active = FALSE
WHERE id = $1;
`

// Список ТОЛЬКО активных
const qListActiveRooms = `
SELECT id, name, is_active
FROM rooms
WHERE is_active = TRUE
ORDER BY id;
`

const qGetRoomByID = `
SELECT id, name, is_active
FROM rooms
WHERE id = $1
`
const qGetRoomByName = `
SELECT id, name, is_active
FROM rooms
WHERE name = $1
`
const qActivateRoom = `
UPDATE rooms
SET is_active = TRUE
WHERE id = $1;
`

// LOGS
const (
	qInsertSoglashenie = `
		INSERT INTO soglasheniya (user_id, user_name, date, doveritel, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	qInsertZapros = `
		INSERT INTO zaprosy (user_id, user_name, date, doveritel, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id;
	`

	qSelectSoglasheniyaByUser = `
		SELECT id, user_id, user_name, date, doveritel, comment, created_at
		FROM soglasheniya
		WHERE user_id = $1
		ORDER BY id DESC
    LIMIT 5;
	`

	qSelectZaprosyByUser = `
		SELECT id, user_id, user_name, date, doveritel, comment, created_at
		FROM zaprosy
		WHERE user_id = $1
		ORDER BY id DESC
    LIMIT 5;
	`

	qSelectSoglashenieByID = `
		SELECT id, user_id, user_name, date, doveritel, comment, created_at
		FROM soglasheniya
		WHERE id = $1;
	`

	qSelectZaprosByID = `
		SELECT id, user_id, user_name, date, doveritel, comment, created_at
		FROM zaprosy
		WHERE id = $1;
	`

	qInsertUser = `
		INSERT INTO users (id, fio)
		VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET fio = EXCLUDED.fio;
	`

	qSelectUserByID = `
		SELECT id, fio, created_at
		FROM users
		WHERE id = $1;
	`
	qSelectSoglasheniyaAfterDate = `
		SELECT id, user_id, user_name, date, doveritel, comment, created_at
		FROM soglasheniya
		WHERE created_at >= $1 AND created_at <= NOW()
		ORDER BY id DESC;
	`

	qSelectZaprosyAfterDate = `
		SELECT id, user_id, user_name, date, doveritel, comment, created_at
		FROM zaprosy
		WHERE created_at >= $1 AND created_at <= NOW()
		ORDER BY id DESC;
	`
)

// docker exec -it db psql -U user -d bookingbot-db -f /tmp/002_logs.up.sql
