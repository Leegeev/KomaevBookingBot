package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

type usecaseforme interface {
	CreateBooking(ctx context.Context, cmd bookingSession) error
	CancelBooking(ctx context.Context, bookingID int64) error
	CheckBookingAndUserID(ctx context.Context, bookingID, userID int64) (bool, error)
	ListUserBookings(ctx context.Context, userID int64) ([]domain.Booking, error)
	ListRoomBookings(ctx context.Context, roomID int64) ([]domain.Booking, error)
	ListRooms(ctx context.Context) ([]domain.Room, error)
	GetRoom(ctx context.Context, roomID int64) (domain.Room, error)
	AdminCreateRoom(ctx context.Context, name string) error
	AdminDeleteRoom(ctx context.Context, roomID int64) error
	// FreeSlots(ctx context.Context, roomID domain.RoomID, day time.Time, step time.Duration) ([]domain.TimeRange, error)
}

// handleMy
// handleBookStart
// handleSchedule
// handleCreateRoom
// handleDeactivateRoom

/*
// handleStart
// handleHelp
*/

/* ---------- /start /help ---------- */

func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /start handler",
			"user", msg.From.UserName,
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
			"err", err)
		return
	}

	h.log.Info("Received /start command",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	msgText := getStartMessageText()

	role, err := h.getRole(ctx, msg.From.ID)
	if err != nil {
		h.log.Error("Failed to get user role", "user_id", msg.From.ID, "err", err)
	} else {
		if role == Administrator || role == Creator {
			msgText += "\n\n" + getAdminStartMessageText()
		}
	}

	escaped := EscapeMarkdownV2(msgText)

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, escaped)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /start message",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", err)
	}
}

func (h *Handler) handleHelp(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /help handler",
			"user", msg.From.UserName,
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
			"err", err)
		return
	}

	h.log.Info("Received /help command",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	msgText := getHelpMessageText()

	role, err := h.getRole(ctx, msg.From.ID)
	if err != nil {
		h.log.Error("Failed to get user role", "user_id", msg.From.ID, "err", err)
	} else {
		if role == Administrator || role == Creator {
			msgText += "\n\n" + getAdminHelpMessageText()
		}
	}

	escaped := EscapeMarkdownV2(msgText)

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, escaped)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /help message",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", err)
	}
}

/* ---------- /my ---------- */
func (h *Handler) handleMy(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /my handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	h.log.Info("Received /my command",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	bookings, err := h.uc.ListUserBookings(ctx, int64(msg.From.ID))
	if err != nil {
		h.log.Error("Failed to list user bookings", "user_id", msg.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка:* `%s`", err.Error()))
		h.reply(msg.Chat.ID, "Возникла ошибка при получении ваших броней.")
		return
	}

	if len(bookings) == 0 {
		h.reply(msg.Chat.ID, "У вас нет будущих броней.")
		return
	}

	text := "*Ваши брони:*"
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(bookings))
	for _, bk := range bookings {
		start := bk.Range.Start.In(h.cfg.OfficeTZ)
		end := bk.Range.End.In(h.cfg.OfficeTZ)
		room, _ := h.uc.GetRoom(ctx, int64(bk.RoomID))

		btnText := fmt.Sprintf("#%s — %s %02d:%02d–%02d:%02d",
			room.Name, start.Format("01-02"),
			start.Hour(), start.Minute(), end.Hour(), end.Minute())

		data := fmt.Sprintf("my:select:%d", bk.ID)

		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	btnText := "Назад"
	data := fmt.Sprintf("my:back")
	btn := tgbotapi.NewInlineKeyboardButtonData(btnText, data)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))

	m := tgbotapi.NewMessage(msg.Chat.ID, EscapeMarkdownV2(text))
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if _, err := h.bot.Send(m); err != nil {
		h.log.Error("Failed to send /my list", "err", err)
	}
}

/* ---------- /rooms ---------- */
func (h *Handler) handleRooms(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /rooms handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	}

	h.log.Info("Received /rooms command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)
	rooms, err := h.uc.ListRooms(ctx)
	if err == domain.ErrNoRoomsAvailable {
		h.log.Warn("No rooms available", "error", err)
		h.reply(msg.Chat.ID, "Нет активных переговорок.")
		return
	} else if err != nil {
		h.log.Error("Failed to list rooms", "error", err)
		h.reply(msg.Chat.ID, "Возникла ошибка при получении списка переговорок: ")
		return
	}
	if len(rooms) == 0 {
		h.reply(msg.Chat.ID, "Активных переговорок нет.")
		return
	}
	var b strings.Builder
	b.WriteString("*Переговорки:*\n")
	for _, r := range rooms {
		fmt.Fprintf(&b, "• %d — %s\n", r.ID, r.Name)
	}
	h.reply(msg.Chat.ID, b.String())
}

/* ---------- /cancel [id] ---------- */

func (h *Handler) handleCancelCommand(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /cancel handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	}
	h.log.Info("Received /cancel command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)
	arg := strings.TrimSpace(msg.CommandArguments())
	/*
		ATTENTION!!
		I SHOULD DO THIS VIA CALLBACK DATA
		ATTENTION!!
		❌❌❌❌❌
	*/
	if arg == "" {
		h.reply(msg.Chat.ID, "Формат: `/cancel <id>` или воспользуйтесь /my и нажмите «Отменить».")
		return
	}
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil || id <= 0 {
		h.reply(msg.Chat.ID, "Некорректный id.")
		return
	}
	// false, domain.ErrNotOwner

	_, err = h.uc.CheckBookingAndUserID(ctx, id, msg.From.ID)
	// если юзер не владелец И не админ
	if err == domain.ErrNotOwner && !h.isAdmin(ctx, int64(msg.From.ID)) {
		h.reply(msg.Chat.ID, "Недостаточно прав для отмены этой брони.")
		return
	}
	// если он админ ИЛИ владелец, если он владелец И админ

	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	// Он либо админ, либо владелец брони, так что можно отменять.
	if err := h.uc.CancelBooking(ctx, id); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Бронь отменена.")
}

/* ---------- /create_room /deactivate_room ---------- */

func (h *Handler) handleCreateRoom(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /CreateRoom handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	}
	h.log.Info("Received /create_room command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)
	if !h.isAdmin(ctx, int64(msg.From.ID)) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}
	name := strings.TrimSpace(msg.CommandArguments())
	if name == "" {
		h.reply(msg.Chat.ID, "Формат: `/create_room <name>`")
		return
	}
	if err := h.uc.AdminCreateRoom(ctx, name); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Комната создана.")
}

func (h *Handler) handleDeactivateRoom(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /DeleteRoom handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	}
	h.log.Info("Received /deactivate_room command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)
	if !h.isAdmin(ctx, int64(msg.From.ID)) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}
	arg := strings.TrimSpace(msg.CommandArguments())
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil || id <= 0 {
		h.reply(msg.Chat.ID, "Формат: `/deactivate_room <room_id>`")
		return
	}
	if err := h.uc.AdminDeleteRoom(ctx, id); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Комната деактивирована.")
}
