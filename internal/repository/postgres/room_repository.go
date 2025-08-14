package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type roomRepositoryPG struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewRoomRepositoryPG(db *sqlx.DB, logger logger.Logger) *roomRepositoryPG {
	return &roomRepositoryPG{db: db, logger: logger}
}

type RoomRepository interface {
	Create(ctx context.Context, r domain.Room) error
	Delete(ctx context.Context, id domain.RoomID) error
	List(ctx context.Context) ([]domain.Room, error)
}
