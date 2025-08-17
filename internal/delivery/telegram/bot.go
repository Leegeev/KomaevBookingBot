package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/internal/usecase"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type Config struct {
	Token       string
	GroupChatID int64          // id вашей беседы для проверок админства
	OfficeTZ    *time.Location // локальная TZ офиса
}

type Handler struct {
	bot *tgbotapi.BotAPI
	cfg Config
	log logger.Logger
	uc  *usecase.BookingService

	// очень простой state для /book: userID -> выбранная комната
	pending map[int64]domain.RoomID
}

func NewHandler(bot *tgbotapi.BotAPI, cfg Config, log logger.Logger, uc *usecase.BookingService) *Handler {
	return &Handler{
		bot:     bot,
		cfg:     cfg,
		log:     log,
		uc:      uc,
		pending: make(map[int64]domain.RoomID),
	}
}

// RunPolling — блокирует до ctx.Done().
func (h *Handler) RunPolling(ctx context.Context) error {
	// sanity check токена/сети
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
	defer func() { // защита от паники, чтобы бот не умер
		if r := recover(); r != nil {
			h.log.Error("panic in handler", "panic", r)
		}
	}()

	if upd.Message != nil {
		msg := upd.Message

		// если пользователь в процессе /book и прислал текст со временем — обработаем
		if !msg.IsCommand() && h.hasPending(msg.From.ID) {
			h.handleBookTime(ctx, msg)
			return
		}

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
			default:
				h.reply(msg.Chat.ID, "Неизвестная команда. Посмотри /help")
			}
			return
		}

		// обычный текст вне сценария
		if strings.TrimSpace(msg.Text) != "" {
			h.reply(msg.Chat.ID, "Не понял. Посмотри /help")
		}
		return
	}

	if upd.CallbackQuery != nil {
		h.handleCallback(ctx, upd.CallbackQuery)
		return
	}
}

/* ---------- small helpers ---------- */

func (h *Handler) reply(chatID int64, text string) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ParseMode = "Markdown"
	_, _ = h.bot.Send(m)
}

func (h *Handler) isAdmin(ctx context.Context, userID int64) bool {
	// если нет группы — считаем всех не-админами
	if h.cfg.GroupChatID == 0 {
		return false
	}
	m, err := h.bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: h.cfg.GroupChatID, UserID: userID,
	})
	if err != nil {
		h.log.Error("GetChatMember failed", "error", err)
		return false
	}
	return m.Status == "creator" || m.Status == "administrator"
}

func (h *Handler) hasPending(userID int64) bool { _, ok := h.pending[userID]; ok = ok; return ok }

// package telegram

// import (
// 	"context"
// 	"fmt"
// 	"strings"

// 	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
// 	"github.com/leegeev/KomaevBookingBot/internal/usecase"
// 	"github.com/leegeev/KomaevBookingBot/pkg/config"
// 	"github.com/leegeev/KomaevBookingBot/pkg/logger"
// )

// func StartBot(ctx context.Context, config config.Telegram, service *usecase.BookingService, logger logger.Logger) error {
// 	bot, err := tgbotapi.NewBotAPI(config.Token)
// 	if err != nil {
// 		logger.Error("Failed to create Telegram bot", "error", err)
// 		return fmt.Errorf("telegram: create bot api: %w", err)
// 	}

// 	// Fail-fast: проверим токен/сетку сразу
// 	if _, err := bot.GetMe(); err != nil {
// 		return fmt.Errorf("telegram: getMe failed: %w", err)
// 	}

// 	logger.Info("Authorized on account", "account", bot.Self.UserName)

// 	// Настройка long polling
// 	u := tgbotapi.NewUpdate(0)
// 	u.Timeout = 30
// 	u.AllowedUpdates = []string{"message", "callback_query"}

// 	updates := bot.GetUpdatesChan(u)
// 	defer bot.StopReceivingUpdates()

// 	// Ежедневный дайджест — отдельная горутина, живёт до ctx.Done().
// 	// Внутри dailyDigestLoop обязательно слушайте ctx и корректно заверщайте sleep/cron.
// 	go dailyDigestLoop(ctx, bot, roomRepo, bookingRepo, logger, config.OfficeTZ)

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			// Корректное завершение по сигналу/отмене
// 			return ctx.Err()

// 		case upd, ok := <-updates:
// 			if !ok {
// 				// Канал апдейтов закрыт библиотекой — выходим, чтобы процесс не «висел»
// 				return fmt.Errorf("telegram: updates channel closed")
// 			}

// 			if upd.Message != nil {
// 				txt := upd.Message.Text

// 				switch {
// 				case strings.HasPrefix(txt, "/start"):
// 					handleStart(ctx, bot, service, logger, upd.Message)

// 				case strings.HasPrefix(txt, "/book"):
// 					handleBooking(ctx, bot, service, logger, upd.Message)

// 				case strings.HasPrefix(txt, "/my"):
// 					handleMy(ctx, bot, service, logger, upd.Message)

// 				case strings.HasPrefix(txt, "/cancel"):
// 					handleCancel(ctx, bot, service, logger, upd.Message)
// 				}

// 			} else if upd.CallbackQuery != nil {
// 				handleCallback(ctx, bot, service, logger, upd.CallbackQuery)
// 			}
// 		}
// 	}
// }

// /*
// 	ready := make(chan struct{})
// 	g.Go(func() error {
// 		return telegram.StartBotWithReady(ctx, cfg.Telegram, roomRepo, userRepo, bookingRepo, log, ready)
// 	})
// 	select {
// 	case <-ready:
// 		log.Info("Telegram bot is ready")
// 	case <-time.After(10 * time.Second):
// 		log.Error("Telegram bot did not become ready in time")
// 		stop()
// 	}
// */
