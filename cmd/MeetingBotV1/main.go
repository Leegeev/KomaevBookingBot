package main

import (
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

// front нужно сделать

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
	logger.Info("Configuration loaded successfully", "config", config)

	// Подключение к БД
	db, err := db.ConnectDBWithRetry(ctx, config.DB, logger)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	// Инициализация репозиториев
	roomRepo := repository.NewRoomRepositoryPG(db, logger)
	bookingRepo := repository.NewBookingRepositoryPG(db, logger)

	// Инициализация сервиса
	service := usecase.NewBookingService(roomRepo, bookingRepo, logger, config.Telegram)

	// TG BOT
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		logger.Error("Failed to init Telegram bot", "error", err)
		return
	}
	h := telegram.NewHandler(bot, config.Telegram, logger, service)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Info("Telegram bot starting...")
		if err := h.RunPolling(ctx); err != nil {
			logger.Error("bot stopped", "error", err)
		}
		logger.Info("Telegram bot stopped")
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		logger.Info("shutting down...")
		bot.StopReceivingUpdates()
		return nil
	})

	if err := g.Wait(); err != nil {
		logger.Error("Service stopped with error", "error", err)
	}
	logger.Info("Service exited cleanly")
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
