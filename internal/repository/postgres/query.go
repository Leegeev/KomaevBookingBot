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
