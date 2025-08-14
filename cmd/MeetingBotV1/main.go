package main

import (
	"os/signal"
	"syscall"
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram"
	db "github.com/leegeev/KomaevBookingBot/internal/infrastructure"
	repository "github.com/leegeev/KomaevBookingBot/internal/repository/postgres"
	"github.com/leegeev/KomaevBookingBot/internal/usecase"
	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
	"golang.org/x/sync/errgroup"

	"context"
	"os"
)

/*
Features:
- Бронирование только в личке
- Авторизация (нужна также роль админа для редактирования расписания брони)
Для админа кнопку возможность добавления/удаления переговорок с фото и описанием
- Ежедневное уведомление (обновляться должно после каждого изменения в расписании или бронировании)
- Хранить думаю стоит не больше недели, потом удалять так как избыточная информация

// кнопки
- /start - приветствие и краткая справка
- /help - полная справка по командам
- /book - начать бронирование переговорки
- /cancel - отменить бронирование
- /rooms - список переговорок
- /mybookings - список моих бронирований
- /admin - админка (только для админов)

TODO:
bot handlers
usecase service
repository implementation
*/

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

	// Подключение к базе данных с повторными попытками
	db, err := db.ConnectDBWithRetry(ctx, config.DB, logger)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	// Инициализация репозиториев
	roomRepo := repository.NewRoomRepositoryPG(db, logger)
	userRepo := repository.NewUserRepositoryPG(db, logger)
	bookingRepo := repository.NewBookingRepositoryPG(db, logger)

	service := usecase.NewBookingService(roomRepo, userRepo, bookingRepo, logger)

	// Запуск бота
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("Telegram bot starting...")
		// ВАЖНО: StartBot должен блокировать до ctx.Done() и возвращать ошибку при фатале.
		if err := telegram.StartBot(ctx, config.Telegram, service, logger); err != nil {
			return err
		}
		logger.Info("Telegram bot stopped")
		return nil
	})
	if err := g.Wait(); err != nil {
		logger.Error("Service stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("Service exited cleanly")
	_ = time.Second // (иногда полезно дать логам долететь; обычно не нужно)
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
