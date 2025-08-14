package main

import (
	"os/signal"
	"syscall"

	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"

	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Initialize logger
	logger := logger.SetupLogger()

	// контекст с отменой для Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupGracefulShutdown(cancel, logger)

	// Load configuration
	config, err := config.LoadConfig(logger)
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		return
	}

	bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
	if err != nil {
		log.Fatal(err)
	}

	// Подключение к базе данных с повторными попытками
	db, err := db.ConnectDBWithRetry(ctx, cfg.DB, logger)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	updates := bot.GetUpdatesChan(u)

	go dailyDigestLoop(ctx, bot, db) // отправка ежедневного статуса

	for upd := range updates {
		if upd.Message != nil {
			switch {
			case strings.HasPrefix(upd.Message.Text, "/start"):
				handleStart(ctx, bot, db, upd.Message)
			case strings.HasPrefix(upd.Message.Text, "/book"):
				startBooking(ctx, bot, db, upd.Message)
			case strings.HasPrefix(upd.Message.Text, "/my"):
				handleMy(ctx, bot, db, upd.Message)
			case strings.HasPrefix(upd.Message.Text, "/cancel"):
				handleCancel(ctx, bot, db, upd.Message)
			}
		} else if upd.CallbackQuery != nil {
			handleCallback(ctx, bot, db, upd.CallbackQuery)
		}
	}
}

func setupGracefulShutdown(cancelFunc context.CancelFunc, logger logger.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", "signal", sig)
		cancelFunc()
	}()
}

// Пример записи брони (упрощённо): вставляем как tstzrange [start, end)
func createBooking(ctx context.Context, db *pgxpool.Pool, roomID int, userID int64, startUTC, endUTC time.Time, note string) error {
	_, err := db.Exec(ctx, `
		INSERT INTO bookings (room_id, created_by, time_range, note)
		VALUES ($1, $2, tstzrange($3, $4, '[)'), $5)
	`, roomID, userID, startUTC, endUTC, note)
	return err // если пересечение — вернётся ошибка из EXCLUDE CONSTRAINT
}
