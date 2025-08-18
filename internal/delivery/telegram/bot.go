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

// Основной хэндлер
type Handler struct {
	bot      *tgbotapi.BotAPI
	cfg      config.Telegram
	log      logger.Logger
	uc       *usecase.BookingService
	bookSess map[int64]*bookingSession // userID -> сессия бронирования
}

func NewHandler(bot *tgbotapi.BotAPI, cfg config.Telegram, log logger.Logger, uc *usecase.BookingService) *Handler {
	return &Handler{
		bot:      bot,
		cfg:      cfg,
		log:      log,
		uc:       uc,
		bookSess: make(map[int64]*bookingSession),
	}
}

// Запуск long-polling. Блокирует до ctx.Done().
func (h *Handler) RunPolling(ctx context.Context) error {
	if _, err := h.bot.GetMe(); err != nil {
		return fmt.Errorf("getMe: %w", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30
	u.AllowedUpdates = []string{"message", "callback_query"}

	updates := h.bot.GetUpdatesChan(u)
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
		}
	}
}

func (h *Handler) dispatch(ctx context.Context, upd tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("panic in telegram handler", "panic", r)
		}
	}()

	if upd.Message != nil {
		msg := upd.Message

		// если в процессе /book — текст не обрабатываем (у нас кнопки)
		if msg.IsCommand() {
			switch msg.Command() {
			case "start":
				h.handleStart(ctx, msg)
			case "help":
				h.handleHelp(ctx, msg)
			case "rooms":
				h.handleRooms(ctx, msg)
			case "my":
				h.handleMy(ctx, msg)
			case "cancel":
				h.handleCancelCommand(ctx, msg)
			case "book":
				h.handleBookStart(ctx, msg)
			case "create_room":
				h.handleCreateRoom(ctx, msg)
			case "deactivate_room":
				h.handleDeactivateRoom(ctx, msg)
			// // TODO:
			// case "schedule":
			// 	h.handleSchedule(ctx, msg)
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
		h.handleCallback(ctx, upd.CallbackQuery)
		return
	}
}

/* ------------ helpers ------------ */

func (h *Handler) reply(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	_, _ = h.bot.Send(m)
}

func (h *Handler) isAdmin(ctx context.Context, userID int64) bool {
	if h.cfg.GroupChatID == 0 {
		return false
	}
	m, err := h.bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: h.cfg.GroupChatID,
			UserID: userID,
		},
	})
	if err != nil {
		h.log.Error("GetChatMember failed", "error", err)
		return false
	}
	return m.Status == "creator" || m.Status == "administrator"
}
