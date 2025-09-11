package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

func (h *Handler) handleSchedule(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /my handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	h.log.Info("Received /book command",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	rooms, err := h.uc.ListRooms(ctx)
	if errors.Is(err, domain.ErrNoRoomsAvailable) {
		h.reply(msg.From.ID, tools.TextBookNoRoomsAvailable.String())
		return
	} else if err != nil {
		h.log.Error("Failed to list rooms", "user_id", msg.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /book:* `%s`", err.Error()))
		h.reply(msg.From.ID, tools.TextBookNoRoomsErr.String())
		return
	}

	for _, room := range rooms {
		var b strings.Builder
		b.WriteString(tools.TextScheduleIntroduction.String() + "\n\n")

		bookings, err := h.uc.ListRoomBookings(ctx, int64(room.ID))
		if err != nil {
			h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /schedule:* `%s`", err.Error()))
			h.log.Error("Failed to list room bookings", "err", err)
			continue
		}

		if len(bookings) == 0 {
			b.WriteString(fmt.Sprintf("*%s*\n_Нет бронирований на ближайшую неделю_\n", room.Name))
		}

		b.WriteString(tools.BuildBookingStr(bookings).String())

		m := tgbotapi.NewMessage(msg.Chat.ID, b.String())
		m.ParseMode = "MarkdownV2"

		if _, err := h.bot.Send(m); err != nil {
			h.log.Error("Failed to handle /book on rooms list", "err", err)
		}
	}
}

// ListRoomBookings
// ListRooms
