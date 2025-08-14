package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type userRepositoryPG struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewUserRepositoryPG(db *sqlx.DB, logger logger.Logger) *userRepositoryPG {
	return &userRepositoryPG{db: db, logger: logger}
}

type UserRepository interface {
	Create(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, userID domain.UserID) error
	IsWhitelisted(ctx context.Context, userID domain.UserID) (bool, error)
}
