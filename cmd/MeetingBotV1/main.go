package main

import (
	"os/signal"
	"syscall"
	"time"

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
	service := usecase.NewBookingService(roomRepo, bookingRepo, logger)

	// TG BOT
	bot, _ := tgbotapi.NewBotAPI(config.Telegram.Token)
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
	if err := g.Wait(); err != nil {
		logger.Error("Service stopped with error", "error", err)
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

/*
	Features:
	- Бронирование только в личке
	- Авторизация (нужна также роль админа для редактирования расписания брони)
	Для админа кнопку возможность добавления/удаления переговорок с фото и описанием
	- Ежедневное уведомление (обновляться должно после каждого изменения в расписании или бронировании)
	- Хранить думаю стоит не больше недели, потом удалять так как избыточная информация

	// кнопки
	/schedule - показать расписание на сегодня

	при создании брони проверять, существует ли комната и активна ли она
	при запросе бронирований выдавать названия комнат, а не ID

	готово:

	- /start - приветствие и краткая справка
	- /help - полная справка по командам
	- /book - начать бронирование переговорки
	- /cancel - отменить свое бронирование (если в беседе у пользователя роль администратора или создателя, он может отменять и чужие брони)
	- /rooms - список переговорок
	- /my - список моих бронирований
	- /create_room
	- /deactivate_room

	TODO:
	bot handlers
	usecase service
	repository implementation

	при отмене бронирования, в тг сообщении нужно в каждую выданную бронь (которую можно отменить) вложить айди брони

	валидация:
	delivery : проверка, что данные есть и они в корректном формате, а пользователь авторизован
	usecase : проверка, что комната существует, что время корректное, что нет пересечений
	repository : проверка, что комната активна (если нужно), что нет пересечений в базе

	delivery только парсит и переводит в UTC;
	usecase создает/валидирует через домен;
	repo хранит, БД окончательно защищает (EXCLUDE).

	комнату лучше не удалять, а просто выключать доступ к ней
	зачем? чтобы не терять историю бронирований
	- для этого в Room добавить поле Active bool
	- при создании брони проверять, что комната активна
	- при получении списка комнат, фильтровать по Active

*/
