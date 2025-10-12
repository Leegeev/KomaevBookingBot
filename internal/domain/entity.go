package domain

import "time"

type (
	RoomID    int64
	UserID    int64
	BookingID int64
	ZaprosID  int64
	SoglID    int64
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
	ID       BookingID
	RoomID   RoomID
	RoomName string // денормализуем для истории
	UserID   UserID
	UserName string
	Range    TimeRange // [start, end) UTC
	Note     string
}

type Soglashenie struct {
	ID        SoglID
	UserID    UserID
	UserName  string
	Doveritel string
	Comment   string
	Date      time.Time
	CreatedAt time.Time
}

type Zapros struct {
	ID        ZaprosID
	UserID    UserID
	UserName  string
	Doveritel string
	Comment   string
	Date      time.Time
	CreatedAt time.Time
}
