package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type bookingRepositoryPG struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewBookingRepositoryPG(db *sqlx.DB, logger logger.Logger) *bookingRepositoryPG {
	return &bookingRepositoryPG{db: db, logger: logger}
}

type BookingRepository interface {
	Create(ctx context.Context, b *domain.Booking) error
	Delete(ctx context.Context, id domain.BookingID) error
	GetByID(ctx context.Context, id domain.BookingID) (domain.Booking, error)

	// Для отображения и проверок
	ListByRoomAndInterval(ctx context.Context, roomID domain.RoomID, fromUTC, toUTC time.Time) ([]domain.Booking, error)
	ListFutureByUser(ctx context.Context, userID domain.UserID, fromUTC time.Time) ([]domain.Booking, error)

	// Быстрая проверка пересечений до вставки (опционально; БД всё равно гарантирует).
	AnyOverlap(ctx context.Context, roomID domain.RoomID, tr domain.TimeRange) (bool, error)

	// Санитарная очистка старых записей.
	DeleteEndedBefore(ctx context.Context, cutoffUTC time.Time) (int64, error)
}
