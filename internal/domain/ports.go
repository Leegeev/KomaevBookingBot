package domain

import (
	"context"
	"time"
)

// Порты (интерфейсы), которые реализуются на уровне инфраструктуры.

// Репозиторий переговорок.
type RoomRepository interface {
	Create(ctx context.Context, r Room) error
	Deactivate(ctx context.Context, id RoomID) error
	Activate(ctx context.Context, id RoomID) error
	List(ctx context.Context) ([]Room, error)
	GetByID(ctx context.Context, id RoomID) (Room, error)
	GetByName(ctx context.Context, name string) (Room, error) // Опционально, если нужно
}

// Репозиторий броней.
type BookingRepository interface {
	// CRUD операции.
	Create(ctx context.Context, b Booking) error
	Delete(ctx context.Context, id BookingID) error
	GetByID(ctx context.Context, id BookingID) (Booking, error)

	// Для отображения и проверок
	ListByRoomAndInterval(ctx context.Context, roomID RoomID, fromUTC, toUTC time.Time) ([]Booking, error)
	ListByUser(ctx context.Context, userID UserID, fromUTC time.Time) ([]Booking, error)

	AnyOverlap(ctx context.Context, roomID RoomID, tr TimeRange) (bool, error)

	// Санитарная очистка старых записей.
	DeleteEndedBefore(ctx context.Context, cutoffUTC time.Time) (int64, error)
}

type LogRepository interface {
	CreateSoglashenie(ctx context.Context, s Soglashenie) (int64, error)
	CreateZapros(ctx context.Context, z Zapros) (int64, error)

	GetSoglasheniyaByUserID(ctx context.Context, userID UserID) ([]Soglashenie, error)
	GetZaprosiByUserID(ctx context.Context, userID UserID) ([]Zapros, error)

	GetSoglashenieByID(ctx context.Context, id int64) (Soglashenie, error)
	GetZaprosByID(ctx context.Context, id int64) (Zapros, error)

	GetSoglasheniyaAfterDate(ctx context.Context, date time.Time) ([]Soglashenie, error)
	GetZaprosiAfterDate(ctx context.Context, date time.Time) ([]Zapros, error)

	GetUser(ctx context.Context, id int64) (User, error)
	CreateUser(ctx context.Context, id int64, FIO string) error
}
