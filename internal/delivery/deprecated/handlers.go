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

// deprecated
/*

// ---------- /rooms ----------
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

// ---------- /cancel [id] ----------

func (h *Handler) handleCancelCommand(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /cancel handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	}
	h.log.Info("Received /cancel command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)
	arg := strings.TrimSpace(msg.CommandArguments())

		// ATTENTION!!
		// I SHOULD DO THIS VIA CALLBACK DATA
		// ATTENTION!!
		// ❌❌❌❌❌

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

*/
