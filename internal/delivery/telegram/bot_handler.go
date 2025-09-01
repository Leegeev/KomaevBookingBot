package telegram

import (
	"context"
	"fmt"
	"strings"
	// "time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/leegeev/KomaevBookingBot/internal/usecase"
	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

// Основной хэндлер
type Handler struct {
	bot       *tgbotapi.BotAPI
	cfg       config.Telegram
	log       logger.Logger
	uc        *usecase.BookingService
	bookStore map[UserID]*bookingSession // userID -> сессия бронирования
	roleCache map[UserID]string          // userID -> роль (user/admin)
}

func NewHandler(bot *tgbotapi.BotAPI, cfg config.Telegram, log logger.Logger, uc *usecase.BookingService) *Handler {
	return &Handler{
		bot: bot,
		cfg: cfg,
		log: log,
		uc:  uc,
		// bookSess: make(map[int64]*bookingSession),
	}
}

// Запуск long-polling. Блокирует до ctx.Done().
func (h *Handler) RunPolling(ctx context.Context) error {
	if _, err := h.bot.GetMe(); err != nil {
		return fmt.Errorf("getMe: %w", err)
	}
	h.bot.Debug = true

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	// updateConfig.AllowedUpdates = []string{"message", "callback_query"}

	updates := h.bot.GetUpdatesChan(updateConfig)
	defer h.bot.StopReceivingUpdates()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case upd, ok := <-updates:
			if !ok {
				return fmt.Errorf("updates channel closed")
			}
			h.dispatch(ctx, upd)
			// go h.dispatch(ctx, upd) // TODO: лимит горутин
		}
	}
}

// Note that panics are a bad way to handle errors. Telegram can
// have service outages or network errors, you should retry sending
// messages or more gracefully handle failures.

func (h *Handler) dispatch(ctx context.Context, upd tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("panic in telegram handler", "panic", r)
		}
	}()

	if upd.Message != nil {
		msg := upd.Message
		if msg.IsCommand() {
			switch msg.Command() {

			// info handlers
			case "start":
				h.handleStart(ctx, msg)
			case "help":
				h.handleHelp(ctx, msg)

			// functional handlers
			case "my":
				h.handleMy(ctx, msg)
			case "book":
				h.handleBook(ctx, msg)
			case "schedule":
				h.handleSchedule(ctx, msg)

			// rooms handlers
			case "create_room":
				h.handleCreateRoom(ctx, msg)
			case "deactivate_room":
				h.handleDeactivateRoom(ctx, msg)

				// case "rooms":
				// h.handleRooms(ctx, msg)
				// case "cancel":
				// h.handleCancelCommand(ctx, msg)
			default:
				h.reply(msg.Chat.ID, "Неизвестная команда. Смотри /help")
			}
			return
		}

		if strings.TrimSpace(msg.Text) != "" {
			h.reply(msg.Chat.ID, "Не понял. Смотри /help")
		}
		return
	}

	if upd.CallbackQuery != nil {
		h.dispatchCallback(ctx, upd.CallbackQuery)
		return
	}
}

func (h *Handler) dispatchCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	if cq == nil || cq.Data == "" {
		h.log.Warn("Empty callback query received")
		return
	}

	parts := strings.Split(cq.Data, ":")
	if len(parts) < 2 {
		h.log.Warn("Invalid callback data format", "data", cq.Data)
		return
	}

	namespace := parts[0]

	switch namespace {
	case "my":
		h.handleMyCallback(ctx, cq)

	case "book":
		h.handleBookCallback(ctx, cq)

	// можно добавить другие пространства имён:
	// case "admin":
	// 	h.handleAdminCallback(ctx, cq)

	default:
		h.log.Warn("Unknown callback namespace", "namespace", namespace, "data", cq.Data)
		h.answerCB(cq, "Неизвестное действие")
	}
}

/* ------------ helpers ------------ */

func (h *Handler) answerCB(cq *tgbotapi.CallbackQuery, text string) {
	cb := tgbotapi.NewCallback(cq.ID, text)

	if _, err := h.bot.Request(cb); err != nil {
		h.log.Error("Failed to answer callback", "err", err, "data", cq.Data)
	}
}

func (h *Handler) reply(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	_, _ = h.bot.Send(m)
}
