package telegram

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/usecase"
	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

func StartBot(ctx context.Context, config config.Telegram, service *usecase.BookingService, logger logger.Logger) error {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		logger.Error("Failed to create Telegram bot", "error", err)
		return fmt.Errorf("telegram: create bot api: %w", err)
	}

	// Fail-fast: проверим токен/сетку сразу
	if _, err := bot.GetMe(); err != nil {
		return fmt.Errorf("telegram: getMe failed: %w", err)
	}

	logger.Info("Authorized on account", "account", bot.Self.UserName)

	// Настройка long polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	u.AllowedUpdates = []string{"message", "callback_query"}

	updates := bot.GetUpdatesChan(u)
	defer bot.StopReceivingUpdates()

	// Ежедневный дайджест — отдельная горутина, живёт до ctx.Done().
	// Внутри dailyDigestLoop обязательно слушайте ctx и корректно заверщайте sleep/cron.
	go dailyDigestLoop(ctx, bot, roomRepo, bookingRepo, logger, config.OfficeTZ)

	for {
		select {
		case <-ctx.Done():
			// Корректное завершение по сигналу/отмене
			return ctx.Err()

		case upd, ok := <-updates:
			if !ok {
				// Канал апдейтов закрыт библиотекой — выходим, чтобы процесс не «висел»
				return fmt.Errorf("telegram: updates channel closed")
			}

			if upd.Message != nil {
				txt := upd.Message.Text

				switch {
				case strings.HasPrefix(txt, "/start"):
					handleStart(ctx, bot, service, logger, upd.Message)

				case strings.HasPrefix(txt, "/book"):
					handleBooking(ctx, bot, service, logger, upd.Message)

				case strings.HasPrefix(txt, "/my"):
					handleMy(ctx, bot, service, logger, upd.Message)

				case strings.HasPrefix(txt, "/cancel"):
					handleCancel(ctx, bot, service, logger, upd.Message)
				}

			} else if upd.CallbackQuery != nil {
				handleCallback(ctx, bot, service, logger, upd.CallbackQuery)
			}
		}
	}
}

/*
	ready := make(chan struct{})
	g.Go(func() error {
		return telegram.StartBotWithReady(ctx, cfg.Telegram, roomRepo, userRepo, bookingRepo, log, ready)
	})
	select {
	case <-ready:
		log.Info("Telegram bot is ready")
	case <-time.After(10 * time.Second):
		log.Error("Telegram bot did not become ready in time")
		stop()
	}
*/
