package tools

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func ParseTimePick(input string) (time.Time, error) {
	trimmed := strings.TrimSpace(input)
	if !strings.Contains(trimmed, ":") {
		trimmed += ":00"
	}

	parsed, err := time.Parse("15:04", trimmed)
	if err != nil {
		return time.Time{}, fmt.Errorf("не удалось разобрать время: %w", err)
	}

	if parsed.Minute() != 0 && parsed.Minute() != 30 {
		return time.Time{}, errors.New("минуты могут быть только 00 или 30")
	}

	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(),
		parsed.Hour(), parsed.Minute(), 0, 0, now.Location()), nil
}

const (
	Creator       = "creator"
	Administrator = "administrator"
	Member        = "member"
)

func CheckRoleIsAdmin(role string) bool {
	return role == Administrator || role == Creator
}

func CheckRoleIsSupported(role string) bool {
	return role == Creator || role == Administrator || role == Member
}
