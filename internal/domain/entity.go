package domain

import "time"

type (
	RoomID    int64
	UserID    int64
	BookingID int64
)

// Сущность комнаты для бронирования.
type Room struct {
	ID       RoomID
	Name     string // уникальное имя, напр. "Переговорка 1"
	IsActive bool   // если false, то комната неактивна и не отображается в списке
}

// Полуинтервал [Start, End): правая граница открыта (стыковка без пересечения).
type TimeRange struct {
	Start time.Time // UTC
	End   time.Time // UTC, > Start
}

// Сущность бронирования комнаты.
type Booking struct {
	ID        BookingID
	RoomID    RoomID
	CreatedBy UserID
	Range     TimeRange // [start, end) UTC
	Note      string
}
