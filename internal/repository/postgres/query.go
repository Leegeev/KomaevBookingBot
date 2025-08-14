package repository

const scheduleQuery = `
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
