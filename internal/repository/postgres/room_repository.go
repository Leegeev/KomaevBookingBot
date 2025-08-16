package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type roomRepositoryPG struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewRoomRepositoryPG(db *sqlx.DB, l logger.Logger) *roomRepositoryPG {
	return &roomRepositoryPG{db: db, logger: l}
}

type roomRow struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// methods

func (r *roomRepositoryPG) Create(ctx context.Context, room domain.Room) (domain.RoomID, error) {
	if room.Name == "" {
		return 0, fmt.Errorf("room name is empty")
	}

	// Хотим убедиться, что вставка прошла — читаем id, но наружу его не вернуть (сигнатура Create не позволяет)
	var newID int64
	if err := r.db.QueryRowxContext(ctx, qInsertRoom, room.Name).Scan(&newID); err != nil {
		// тут можно дополнительно замапить уникальное имя на доменную ошибку, если нужно
		// (код PG: 23505). Иначе — отдать как есть.
		return 0, fmt.Errorf("failed to create room: %w", err)
	}
	return domain.RoomID(newID), nil
}

func (r *roomRepositoryPG) Delete(ctx context.Context, id domain.RoomID) error {
	res, err := r.db.ExecContext(ctx, qDeleteRoom, int64(id))
	if err != nil {
		return err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		// маппим в доменную «не найдено», если она у тебя есть
		// (раньше мы вводили domain.ErrRoomNotFound; если нет — можно вернуть sql.ErrNoRows)
		if domain.ErrRoomNotFound != nil {
			return domain.ErrRoomNotFound
		}
		return sql.ErrNoRows
	}
	return nil
}

func (r *roomRepositoryPG) List(ctx context.Context) ([]domain.Room, error) {
	var rows []roomRow
	if err := r.db.SelectContext(ctx, &rows, qListRooms); err != nil {
		return nil, err
	}
	rooms := make([]domain.Room, 0, len(rows))
	for _, rr := range rows {
		rooms = append(rooms, roomRowToDomain(rr))
	}
	return rooms, nil
}

// helper functions

func roomRowToDomain(rr roomRow) domain.Room {
	return domain.Room{
		ID:   domain.RoomID(rr.ID),
		Name: rr.Name,
	}
}
