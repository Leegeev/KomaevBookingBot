package telegram

import (
	"context"
	"fmt"
	"strings"
	"sync"

	// "time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/leegeev/KomaevBookingBot/internal/delivery/notifier"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
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
	sessions  *tools.SessionsStore // userID -> сессия бронирования
	roleCache *RoleCache           // userID -> роль (user/admin)

	messageID int64
	msgMu     sync.Mutex

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
		roleCache:        NewRoleCache(cfg.RoleCacheTTL),
		messageID:        0,
		msgMu:            sync.Mutex{},
		commandHandlers:  make(map[string]func(ctx context.Context, msg *tgbotapi.Message)),
		callbackHandlers: make(map[string]func(ctx context.Context, cq *tgbotapi.CallbackQuery)),
	}
}

// Запуск long-polling. Блокирует до ctx.Done().
func (h *Handler) RunPolling(ctx context.Context) error {
	if _, err := h.bot.GetMe(); err != nil {
		return fmt.Errorf("getMe: %w", err)
	}
	h.bot.Debug = false

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	h.registerRoutes()
	updateConfig.AllowedUpdates = []string{"message", "callback_query"}

	n := notifier.New(h.log, h.cfg.OfficeTZ)
	err := n.AddJob(ctx, h.cfg.NotifierConfig, h.DailySchedule)
	if err != nil {
		h.log.Error("failed to add job", "err", err)
	}
	n.Start(ctx)

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

func (h *Handler) dispatch(ctx context.Context, upd tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			h.log.Error("panic in telegram handler", "panic", r)
		}
	}()

	if upd.Message != nil && upd.Message.IsCommand() && upd.Message.Command() == "register" {
		h.handleRegisterFromAdmin(ctx, upd.Message)
		return
	}

	if err := h.checkSupported(ctx, upd); err != nil {
		h.reply(upd.FromChat().ChatConfig().ChatID, err.Error())
		return
	}

	if upd.Message.Chat.Type != "private" {
		return
	}

	if upd.Message != nil && upd.Message.IsCommand() {
		h.log.Info("Received command",
			"user", upd.Message.From.UserName,
			"user_id", upd.Message.From.ID,
			"chat_id", upd.Message.Chat.ID,
			"command", upd.Message.Command(),
		)
		cmd := upd.Message.Command()
		if handler, ok := h.commandHandlers[cmd]; ok {
			handler(ctx, upd.Message)
		} else {
			h.reply(upd.Message.Chat.ID, "Неизвестная команда. Смотри /help")
		}
		return
	}

	if upd.CallbackQuery != nil {
		h.log.Info("Received callback",
			"user", upd.CallbackQuery.From.UserName,
			"user_id", upd.CallbackQuery.From.ID,
			"chat_id", upd.CallbackQuery.Message.Chat.ID,
			"data", upd.CallbackQuery.Data,
		)
		cb := strings.Split(upd.CallbackQuery.Data, ":")
		prefix := cb[0] + ":" + cb[1]
		if handler, ok := h.callbackHandlers[prefix]; ok {
			handler(ctx, upd.CallbackQuery)
		}
		return
	}

	if upd.Message != nil {
		h.log.Info("Received message",
			"user", upd.Message.From.UserName,
			"user_id", upd.Message.From.ID,
			"chat_id", upd.Message.Chat.ID,
			"text", upd.Message.Text,
		)
		if handler, ok := h.commandHandlers[upd.Message.Text]; ok {
			handler(ctx, upd.Message)
			return
		}

		sess := h.sessions.Get(upd.Message.From.ID)
		switch {
		case sess == nil:
			h.reply(upd.Message.Chat.ID, "Необработанный ввод. Смотри /help")
			return
		case sess.BookState == tools.BookStateChoosingStartTime:
			h.handleBookTimepick(ctx, upd.Message)
			return
		case sess.BookState == tools.StateProccessingRoomCreation:
			h.handleCreateRoomProcessing(ctx, upd.Message)
			return
		default:
			h.reply(upd.Message.Chat.ID, "Сессия не найдена. Смотри /help")
			return
		}
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

func (h *Handler) registerRoutes() {
	// commands
	h.commandHandlers["start"] = h.handleStart
	h.commandHandlers["help"] = h.handleHelp
	h.commandHandlers["my"] = h.handleMy
	h.commandHandlers["book"] = h.handleBook
	h.commandHandlers["schedule"] = h.handleSchedule
	h.commandHandlers["create_room"] = h.handleCreateRoom
	h.commandHandlers["deactivate_room"] = h.handleDeactivateRoom
	h.commandHandlers["register"] = h.handleRegisterFromAdmin

	// Text commands
	h.commandHandlers[tools.TextMainBookButton] = h.handleBook
	h.commandHandlers[tools.TextMainMyButton] = h.handleMy
	h.commandHandlers[tools.TextMainScheduleButton] = h.handleSchedule
	h.commandHandlers[tools.TextMainCreateRoomButton] = h.handleCreateRoom
	h.commandHandlers[tools.TextMainDeleteRoomButton] = h.handleDeactivateRoom
	h.commandHandlers[tools.TextMainHelpButton] = h.handleHelp

	// callbacks
	// BOOK
	h.callbackHandlers["book:list"] = h.handleBookList
	h.callbackHandlers["book:calendar"] = h.handleBookCalendar
	h.callbackHandlers["book:calendar_nav"] = h.handleBookCalendarNavigation // book:calendar_nav:-1
	h.callbackHandlers["book:duration"] = h.handleBookDuration
	h.callbackHandlers["book:confirm"] = h.handleBookConfirm

	h.callbackHandlers["book:list_back"] = h.handleBookListBack
	h.callbackHandlers["book:calendar_back"] = h.handleBookCalendarBack
	h.callbackHandlers["book:timepick_back"] = h.handleBookTimepickBack
	h.callbackHandlers["book:duration_back"] = h.handleBookDurationBack
	h.callbackHandlers["book:confirm_back"] = h.handleBookConfirmBack

	// MY
	h.callbackHandlers["my:list"] = h.handleMyList
	h.callbackHandlers["my:back"] = h.handleMyBack
	h.callbackHandlers["my:cancel"] = h.handleMyCancel
	h.callbackHandlers["my:list_back"] = h.handleMyListBack

	// no:op
	h.callbackHandlers["no:op"] = func(ctx context.Context, cq *tgbotapi.CallbackQuery) {
		h.answerCB(cq, "")
	}

	// ROOMS
	h.callbackHandlers["deactivate_room:list"] = h.handleDeactivateList
	h.callbackHandlers["deactivate_room:list_back"] = h.handleDeactivateListBack
	h.callbackHandlers["deactivate:confirm"] = h.handleDeactivateConfirm
	h.callbackHandlers["deactivate:confirm_cancel"] = h.handleConfirmCancel
	h.callbackHandlers["deactivate:confirm_back"] = h.handleDeactivateConfirmBack

	// h.callbackHandlers["my:reschedule"] = h.handleMyReschedule
}
