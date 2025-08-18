package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/jmoiron/sqlx"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

// ConnectDBWithRetry подключается к PostgreSQL с указанным количеством повторных попыток.
// Использует context с таймаутом для ping. Возвращает подключение *sqlx.DB или ошибку.
func ConnectDBWithRetry(ctx context.Context, cfg config.DB, logger logger.Logger) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error
	dsn := cfg.DSN()
	logger.Info("DSN loaded", "dsn", dsn)
	for attempt := 1; attempt <= cfg.RetryCount; attempt++ {
		db, err = sqlx.Open("pgx", dsn)
		if err != nil {
			logger.Warn("Failed to open DB connection", "attempt", attempt, "error", err)
		} else {
			pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			if pingErr := db.PingContext(pingCtx); pingErr == nil {
				logger.Info("Database connection established", "attempt", attempt)
				return db, nil
			} else {
				if errors.Is(pingErr, context.DeadlineExceeded) || errors.Is(pingErr, context.Canceled) {
					logger.Warn("Ping context deadline exceeded", "attempt", attempt, "error", pingErr)
					return nil, fmt.Errorf("%w, %v", domain.ErrDBConnectionFailed, pingErr)
				}
				logger.Warn("Failed to ping DB", "attempt", attempt, "error", pingErr)
				_ = db.Close()
			}
		}

		time.Sleep(cfg.RetryDelay * time.Duration(attempt)) // Увеличиваем задержку между попытками
	}
	logger.Error("Could not connect to DB after %d attempts, last error: %v", cfg.RetryCount, err)
	return nil, fmt.Errorf("%w, %v", domain.ErrDBConnectionFailed, err)
}
