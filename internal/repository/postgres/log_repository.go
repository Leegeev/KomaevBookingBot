package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
	"time"
)

// Структура репозитория
type logRepositoryPG struct {
	db     *sqlx.DB
	logger logger.Logger
}

// Конструктор
func NewLogRepositoryPG(db *sqlx.DB, logger logger.Logger) *logRepositoryPG {
	return &logRepositoryPG{db: db, logger: logger}
}

// Структуры для работы с БД
type soglashenieRow struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	UserName  string    `db:"user_name"`
	Date      time.Time `db:"date"`
	Doveritel string    `db:"doveritel"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
}

type zaprosRow struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	UserName  string    `db:"user_name"`
	Date      time.Time `db:"date"`
	Doveritel string    `db:"doveritel"`
	Comment   string    `db:"comment"`
	CreatedAt time.Time `db:"created_at"`
}

// ────────────────────────────────
//         Create
// ────────────────────────────────

func (r *logRepositoryPG) CreateSoglashenie(ctx context.Context, s domain.Soglashenie) (int64, error) {
	var newID int64
	err := r.db.QueryRowxContext(ctx, qInsertSoglashenie,
		int64(s.UserID), s.UserName, s.Date, s.Doveritel, s.Comment,
	).Scan(&newID)

	if err != nil {
		return 0, mapPgErr(err)
	}
	return newID, nil
}

func (r *logRepositoryPG) CreateZapros(ctx context.Context, z domain.Zapros) (int64, error) {
	var newID int64
	err := r.db.QueryRowxContext(ctx, qInsertZapros,
		int64(z.UserID), z.UserName, z.Date, z.Doveritel, z.Comment,
	).Scan(&newID)

	if err != nil {
		return 0, mapPgErr(err)
	}
	return newID, nil
}

// ────────────────────────────────
//         Get by UserID
// ────────────────────────────────

func (r *logRepositoryPG) GetSoglasheniyaByUserID(ctx context.Context, userID domain.UserID) ([]domain.Soglashenie, error) {
	var rows []soglashenieRow
	if err := r.db.SelectContext(ctx, &rows, qSelectSoglasheniyaByUser, int64(userID)); err != nil {
		return nil, err
	}

	out := make([]domain.Soglashenie, 0, len(rows))
	for _, row := range rows {
		out = append(out, soglashenieRowToDomain(row))
	}
	return out, nil
}

func (r *logRepositoryPG) GetZaprosiByUserID(ctx context.Context, userID domain.UserID) ([]domain.Zapros, error) {
	var rows []zaprosRow
	if err := r.db.SelectContext(ctx, &rows, qSelectZaprosyByUser, int64(userID)); err != nil {
		return nil, err
	}

	out := make([]domain.Zapros, 0, len(rows))
	for _, row := range rows {
		out = append(out, zaprosRowToDomain(row))
	}
	return out, nil
}

// ────────────────────────────────
//         Get by ID
// ────────────────────────────────

func (r *logRepositoryPG) GetSoglashenieByID(ctx context.Context, id int64) (domain.Soglashenie, error) {
	var row soglashenieRow
	if err := r.db.GetContext(ctx, &row, qSelectSoglashenieByID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Soglashenie{}, domain.ErrRecordNotFound
		}
		return domain.Soglashenie{}, err
	}
	return soglashenieRowToDomain(row), nil
}

func (r *logRepositoryPG) GetZaprosByID(ctx context.Context, id int64) (domain.Zapros, error) {
	var row zaprosRow
	if err := r.db.GetContext(ctx, &row, qSelectZaprosByID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Zapros{}, domain.ErrRecordNotFound
		}
		return domain.Zapros{}, err
	}
	return zaprosRowToDomain(row), nil
}

// ────────────────────────────────
//         Helpers
// ────────────────────────────────

func soglashenieRowToDomain(r soglashenieRow) domain.Soglashenie {
	return domain.Soglashenie{
		ID:        domain.SoglID(r.ID),
		UserID:    domain.UserID(r.UserID),
		UserName:  r.UserName,
		Date:      r.Date,
		Doveritel: r.Doveritel,
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt,
	}
}

func zaprosRowToDomain(r zaprosRow) domain.Zapros {
	return domain.Zapros{
		ID:        domain.ZaprosID(r.ID),
		UserID:    domain.UserID(r.UserID),
		UserName:  r.UserName,
		Date:      r.Date,
		Doveritel: r.Doveritel,
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt,
	}
}

func mapPgErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return fmt.Errorf("duplicate: %w", err)
		case "23503": // foreign_key_violation
			return fmt.Errorf("invalid reference: %w", err)
		}
	}
	return err
}
