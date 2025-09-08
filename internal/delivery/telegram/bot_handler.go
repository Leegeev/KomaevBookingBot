package telegram

import (
	"context"
	"fmt"
	"strings"

	// "time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
	"github.com/leegeev/KomaevBookingBot/internal/usecase"
	"github.com/leegeev/KomaevBookingBot/pkg/config"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

// Основной хэндлер
type Handler struct {
	bot      *tgbotapi.BotAPI
	cfg      config.Telegram
	log      logger.Logger
	uc       *usecase.BookingService
	sessions *tools.SessionsStore // userID -> сессия бронирования
	// roleCache map[UserID]string    // userID -> роль (user/admin)

	commandHandlers  map[string]func(ctx context.Context, msg *tgbotapi.Message)
	callbackHandlers map[string]func(ctx context.Context, cq *tgbotapi.CallbackQuery)
}

func NewHandler(bot *tgbotapi.BotAPI, cfg config.Telegram, log logger.Logger, uc *usecase.BookingService) *Handler {
	return &Handler{
		bot:              bot,
		cfg:              cfg,
		log:              log,
		uc:               uc,
		sessions:         tools.NewSessionStore(),
		commandHandlers:  make(map[string]func(ctx context.Context, msg *tgbotapi.Message)),
		callbackHandlers: make(map[string]func(ctx context.Context, cq *tgbotapi.CallbackQuery)),
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
	h.registerRoutes()
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

func (h *Handler) registerRoutes() {
	// commands
	h.commandHandlers["start"] = h.handleStart
	h.commandHandlers["help"] = h.handleHelp
	h.commandHandlers["my"] = h.handleMy
	h.commandHandlers["book"] = h.handleBook
	h.commandHandlers["schedule"] = h.handleSchedule
	h.commandHandlers["create_room"] = h.handleCreateRoom
	h.commandHandlers["deactivate_room"] = h.handleDeactivateRoom

	// callbacks
	// h.commandHandlers["my"] = h.handleMyCallback
	h.callbackHandlers["book:list"] = h.handleBookList
	h.callbackHandlers["book:calendar"] = h.handleBookCalendar
	h.callbackHandlers["book:duration"] = h.handleBookDuration
	h.callbackHandlers["book:confirm"] = h.handleBookConfirm

	h.callbackHandlers["book:list_back"] = h.handleBookListBack
	h.callbackHandlers["book:calendar_back"] = h.handleBookCalendarBack
	h.callbackHandlers["book:timepick_back"] = h.handleBookTimepickBack
	h.callbackHandlers["book:duration_back"] = h.handleBookDurationBack
	h.callbackHandlers["book:confirm_back"] = h.handleBookConfirmBack

	// h.commandHandlers["create_room"] = h.handleCreateRoomCallback
	// h.commandHandlers["deactivate_room"] = h.handleDeactivateRoomCallback
}

func (h *Handler) dispatch(ctx context.Context, upd tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("panic in telegram handler", "panic", r)
		}
	}()

	if upd.Message.IsCommand() {
		cmd := upd.Message.Command()
		if handler, ok := h.commandHandlers[cmd]; ok {
			handler(ctx, upd.Message)
		} else {
			h.reply(upd.Message.Chat.ID, "Неизвестная команда. Смотри /help")
		}
		return
	}

	if upd.CallbackQuery != nil {
		cb := strings.Split(upd.CallbackQuery.Data, ":")
		prefix := cb[0]
		if handler, ok := h.callbackHandlers[prefix]; ok {
			handler(ctx, upd.CallbackQuery)
		}
		return
	}

	if upd.Message != nil {
		//check if it is time input
		// else
		return
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
