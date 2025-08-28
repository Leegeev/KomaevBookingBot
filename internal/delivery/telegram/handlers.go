package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

type bookingSession struct {
	RoomID   int64
	DayLocal time.Time // полночь локального дня (OfficeTZ)

	StartH, StartM *int
	EndH, EndM     *int
}

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

/* ---------- /start /help ---------- */

func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /start handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
	}

	h.log.Info("Received /start command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)

	text := "*Привет\\! Я бот бронирования переговорок\\.*\n\n" +
		"• /book — забронировать\n" +
		"• /my — мои брони\n" +
		"• /rooms — список переговорок\n" +
		"• /cancel — отменить бронь\n" +
		"• /help — справка\n" +
		"• /create\\_room — создать переговорку \\(админ\\)\n" +
		"• /deactivate\\_room — деактивировать переговорку \\(админ\\)"

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /start message", "error", err)
	}
}

func (h *Handler) handleHelp(ctx context.Context, msg *tgbotapi.Message) {
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /help handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
	}

	h.log.Info("Received /help command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)

	text := "*Команды:*\n" +
		"• /book — выбрать переговорку → день → время начала/окончания кнопками\n" +
		"• /my — показать ваши будущие брони \\(кнопки «Отменить»\\)\n" +
		"• /rooms — показать активные переговорки\n" +
		"• /cancel `<id>` — отменить по номеру\n" +
		"• /create\\_room `<name>` — создать переговорку \\(админ\\)\n" +
		"• /deactivate\\_room `<room_id>` — выключить переговорку \\(админ\\)"

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, text)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /help message", "error", err)
	}
}

/* ---------- /rooms ---------- */
func (h *Handler) handleRooms(ctx context.Context, msg *tgbotapi.Message) {
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /rooms handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
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

/* ---------- /my ---------- */

func (h *Handler) handleMy(ctx context.Context, msg *tgbotapi.Message) {
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /my handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
	}

	h.log.Info("Received /my command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)
	bookings, err := h.uc.ListUserBookings(ctx, int64(msg.From.ID))
	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	if len(bookings) == 0 {
		h.reply(msg.Chat.ID, "У вас нет будущих броней.")
		return
	}

	kbRows := [][]tgbotapi.InlineKeyboardButton{}
	var b strings.Builder
	b.WriteString("*Ваши брони:*\n")
	for _, bk := range bookings {
		start := bk.Range.Start.In(h.cfg.OfficeTZ)
		end := bk.Range.End.In(h.cfg.OfficeTZ)
		roomInfo, _ := h.uc.GetRoom(ctx, int64(bk.RoomID))
		fmt.Fprintf(&b, "• #%s — %s %02d:%02d–%02d:%02d\n",
			roomInfo.Name, start.Format("2006-01-02"), start.Hour(), start.Minute(), end.Hour(), end.Minute())

		cb := tgbotapi.NewInlineKeyboardButtonData("Отменить", fmt.Sprintf("c:%d", bk.ID))
		kbRows = append(kbRows, tgbotapi.NewInlineKeyboardRow(cb))
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, b.String())
	m.ParseMode = "Markdown"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kbRows...)
	_, _ = h.bot.Send(m)
}

/* ---------- /cancel [id] ---------- */

func (h *Handler) handleCancelCommand(ctx context.Context, msg *tgbotapi.Message) {
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /cancel handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
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
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /CreateRoom handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
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
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /DeleteRoom handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
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
