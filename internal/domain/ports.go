package domain

import (
	"context"
	"time"
)

// Порты (интерфейсы), которые реализуются на уровне инфраструктуры.

// Репозиторий переговорок.
type RoomRepository interface {
	Create(ctx context.Context, r Room) error
	Delete(ctx context.Context, id RoomID) error
	List(ctx context.Context) ([]Room, error)
}

// Репозиторий пользователей (whitelist).
// Если getChatMember работает как надо, то можно и не хранить пользователей в БД.
type UserRepository interface {
	Add(ctx context.Context, u User) error
	Delete(ctx context.Context, userID UserID) error
	IsWhitelisted(ctx context.Context, userID UserID) (bool, error)
}

// Репозиторий броней.
type BookingRepository interface {
	Create(ctx context.Context, b *Booking) error
	Delete(ctx context.Context, id BookingID) error
	GetByID(ctx context.Context, id BookingID) (Booking, error)

	// Для отображения и проверок
	ListByRoomAndInterval(ctx context.Context, roomID RoomID, fromUTC, toUTC time.Time) ([]Booking, error)
	ListFutureByUser(ctx context.Context, userID UserID, fromUTC time.Time) ([]Booking, error)

	// Быстрая проверка пересечений до вставки (опционально; БД всё равно гарантирует).
	AnyOverlap(ctx context.Context, roomID RoomID, tr TimeRange) (bool, error)

	// Санитарная очистка старых записей.
	DeleteEndedBefore(ctx context.Context, cutoffUTC time.Time) (int64, error)
}

// // Публикатор доменных событий (можно замокать).
// type EventPublisher interface {
// 	Publish(ctx context.Context, ev Event) error
// }

// // Часы для тестируемости.
// type Clock interface {
// 	NowUTC() time.Time
// }

// type realClock struct{}

// func (realClock) NowUTC() time.Time { return time.Now().UTC() }
// func SystemClock() Clock            { return realClock{} }
