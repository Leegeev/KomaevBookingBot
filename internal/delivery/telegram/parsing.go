package telegram

import (
	"errors"
	"strings"
	"time"
)

// Ожидаем: "YYYY-MM-DD HH:MM-HH:MM[; комментарий]"
// Конвертируем в UTC с учётом officeTZ.
func parseDateTimeRange(input string, officeTZ *time.Location) (time.Time, time.Time, string, error) {
	parts := strings.SplitN(input, ";", 2)
	rangePart := strings.TrimSpace(parts[0])
	note := ""
	if len(parts) == 2 {
		note = strings.TrimSpace(parts[1])
	}

	// сплит даты и диапазона
	spaceIdx := strings.IndexByte(rangePart, ' ')
	if spaceIdx < 0 {
		return time.Time{}, time.Time{}, "", errors.New("bad format")
	}
	dateStr := rangePart[:spaceIdx]
	hhmm := strings.TrimSpace(rangePart[spaceIdx+1:])

	// разбираем времена
	times := strings.Split(hhmm, "-")
	if len(times) != 2 {
		return time.Time{}, time.Time{}, "", errors.New("bad time range")
	}

	// парсим локальные времена
	day, err := time.ParseInLocation("2006-01-02", dateStr, officeTZ)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}
	startHHMM, endHHMM := strings.TrimSpace(times[0]), strings.TrimSpace(times[1])

	sh, sm, err := parseHHMM(startHHMM)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}
	eh, em, err := parseHHMM(endHHMM)
	if err != nil {
		return time.Time{}, time.Time{}, "", err
	}

	startLocal := time.Date(day.Year(), day.Month(), day.Day(), sh, sm, 0, 0, officeTZ)
	endLocal := time.Date(day.Year(), day.Month(), day.Day(), eh, em, 0, 0, officeTZ)

	return startLocal.UTC(), endLocal.UTC(), note, nil
}

func parseHHMM(s string) (int, int, error) {
	if len(s) != 5 || s[2] != ':' {
		return 0, 0, errors.New("bad HH:MM")
	}
	hh, mm := (int(s[0]-'0')*10 + int(s[1]-'0')), (int(s[3]-'0')*10 + int(s[4]-'0'))
	if hh < 0 || hh > 23 || mm < 0 || mm > 59 {
		return 0, 0, errors.New("bad HH or MM")
	}
	return hh, mm, nil
}
