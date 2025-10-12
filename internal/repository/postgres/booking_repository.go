package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type bookingRepositoryPG struct {
	db     *sqlx.DB
	logger logger.Logger
}

type bookingRow struct {
	ID        int64     `db:"id"`
	RoomID    int64     `db:"room_id"`
	RoomName  string    `db:"room_name"`
	UserID    int64     `db:"user_id"`
	UserName  string    `db:"user_name"`
	StartUTC  time.Time `db:"start_utc"`
	EndUTC    time.Time `db:"end_utc"`
	CreatedAt time.Time `db:"created_at"`
}

func NewBookingRepositoryPG(db *sqlx.DB, logger logger.Logger) *bookingRepositoryPG {
	return &bookingRepositoryPG{db: db, logger: logger}
}

func (r *bookingRepositoryPG) Create(ctx context.Context, b domain.Booking) error {
	start := b.Range.Start
	end := b.Range.End

	var newID int64

	err := r.db.QueryRowxContext(
		ctx,
		qInsertBooking,
		b.RoomID,
		b.RoomName,
		b.UserID,
		b.UserName,
		start,
		end,
	).Scan(&newID)
	if err != nil {
		return mapPgOverlapErr(err)
	}

	// b.ID = domain.BookingID(newID)
	return nil
}

func (r *bookingRepositoryPG) Delete(ctx context.Context, id domain.BookingID) error {
	res, err := r.db.ExecContext(ctx, qDeleteByID, int64(id))
	if err != nil {
		return err
	}

	aff, _ := res.RowsAffected()
	if aff == 0 {
		return domain.ErrBookingNotFound
	}
	return nil
}

func (r *bookingRepositoryPG) GetByID(ctx context.Context, id domain.BookingID) (domain.Booking, error) {
	var br bookingRow
	if err := r.db.GetContext(ctx, &br, qSelectByID, int64(id)); err != nil {
		return domain.Booking{}, err
	}
	return bookingRowToDomain(br)
}

func (r *bookingRepositoryPG) ListByRoomAndInterval(ctx context.Context, roomID domain.RoomID, fromUTC, toUTC time.Time) ([]domain.Booking, error) {
	var rows []bookingRow
	if err := r.db.SelectContext(ctx, &rows, qListByRoomAndInterval, int64(roomID), fromUTC, toUTC); err != nil {
		return nil, err
	}
	out := make([]domain.Booking, 0, len(rows))
	for _, br := range rows {
		b, err := bookingRowToDomain(br)
		if err != nil {
			return nil, fmt.Errorf("bookingRowToDomain: %w", err)
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *bookingRepositoryPG) ListByUser(ctx context.Context, userID domain.UserID, fromUTC time.Time) ([]domain.Booking, error) {
	var rows []bookingRow
	if err := r.db.SelectContext(ctx, &rows, qListByUser, int64(userID), fromUTC); err != nil {
		return nil, err
	}
	out := make([]domain.Booking, 0, len(rows))
	for _, br := range rows {
		b, err := bookingRowToDomain(br)
		if err != nil {
			return nil, fmt.Errorf("bookingRowToDomain: %w", err)
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *bookingRepositoryPG) AnyOverlap(ctx context.Context, roomID domain.RoomID, tr domain.TimeRange) (bool, error) {
	var has bool
	if err := r.db.GetContext(ctx, &has, qAnyOverlap, int64(roomID), tr.Start, tr.End); err != nil {
		return false, err
	}
	return has, nil
}

func (r *bookingRepositoryPG) DeleteEndedBefore(ctx context.Context, cutoffUTC time.Time) (int64, error) {
	res, err := r.db.ExecContext(ctx, qDeleteEndedBefore, cutoffUTC)
	if err != nil {
		return 0, err
	}
	aff, _ := res.RowsAffected()
	return aff, nil
}

// helper functions
func bookingRowToDomain(br bookingRow) (domain.Booking, error) {
	tr, err := domain.NewTimeRange(br.StartUTC, br.EndUTC)
	if err != nil {
		return domain.Booking{}, err
	}

	return domain.Booking{
		ID:       domain.BookingID(br.ID),
		RoomID:   domain.RoomID(br.RoomID),
		RoomName: br.RoomName,
		UserID:   domain.UserID(br.UserID),
		UserName: br.UserName,
		Range:    tr,
		// CreatedAt: br.CreatedAt.UTC(),
	}, nil
}

func mapPgOverlapErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23P01" { // exclusion_violation
		return domain.ErrOverlapsExisting
	}
	return err
}

// type BookingRepository interface {
// 	Create(ctx context.Context, b *domain.Booking) error
// 	Delete(ctx context.Context, id domain.BookingID) error
// 	GetByID(ctx context.Context, id domain.BookingID) (domain.Booking, error)

// 	// Для отображения и проверок
// 	ListByRoomAndInterval(ctx context.Context, roomID domain.RoomID, fromUTC, toUTC time.Time) ([]domain.Booking, error)
// 	ListFutureByUser(ctx context.Context, userID domain.UserID, fromUTC time.Time) ([]domain.Booking, error)

// 	// Быстрая проверка пересечений до вставки (опционально; БД всё равно гарантирует).
// 	AnyOverlap(ctx context.Context, roomID domain.RoomID, tr domain.TimeRange) (bool, error)

// 	// Санитарная очистка старых записей.
// 	DeleteEndedBefore(ctx context.Context, cutoffUTC time.Time) (int64, error)
// }
