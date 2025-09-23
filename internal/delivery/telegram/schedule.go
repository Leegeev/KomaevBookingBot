package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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
		schedule := h.scheduleBuilder(ctx, room, time.Now().Add(time.Hour*24*7))
		if schedule == "" {
			h.reply(msg.Chat.ID, tools.TextScheduleError.String())
		}
		m := tgbotapi.NewMessage(msg.Chat.ID, schedule)
		m.ParseMode = "MarkdownV2"

		if _, err := h.bot.Send(m); err != nil {
			h.log.Error("Failed to handle /book on rooms list", "err", err)
		}
	}
}

func (h *Handler) scheduleBuilder(ctx context.Context, room domain.Room, end time.Time) string {
	var b strings.Builder
	b.WriteString(tools.TextScheduleIntroduction.String() + "\n\n")

	bookings, err := h.uc.ListRoomBookings(ctx, int64(room.ID), end)
	if err != nil {
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /schedule:* `%s`", err.Error()))
		h.log.Error("Failed to list room bookings", "err", err)
		return ""
	}

	if len(bookings) == 0 {
		b.WriteString(fmt.Sprintf("*%s*\n_Нет бронирований на ближайшую неделю_\n", room.Name))
	}

	b.WriteString(tools.BuildBookingStr(bookings).String())
	return b.String()
}

func (h *Handler) buildTodaySchedule() string {
	var b strings.Builder
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Minute*5))
	defer cancel()

	rooms, err := h.uc.ListRooms(ctx)
	if err != nil {
		h.log.Error("failed to list rooms for schedule", "err", err)
		h.notifyAdmin("error in BuildTodaySchedule when listing rooms: " + err.Error())
		return "Не удалось получить список переговорок"
	}
	if len(rooms) == 0 {
		return "На данный момент переговорок нет"
	}

	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, h.cfg.OfficeTZ)
	endOfDay := startOfDay.Add(24 * time.Hour)

	b.WriteString(
		tools.EscapeMarkdownV2(
			fmt.Sprintf(tools.NotifierString, today.Format("02.01.2006")),
		),
	)

	for _, room := range rooms {
		bookings, err := h.uc.ListRoomBookings(ctx, int64(room.ID), endOfDay)
		if err != nil {
			h.log.Error("failed to get bookings", "room", room.Name, "err", err)
			continue
		}

		if len(bookings) == 0 {
			b.WriteString(fmt.Sprintf("*%s*\n_Свободна весь день_\n\n",
				tools.EscapeMarkdownV2(room.Name),
			))
			continue
		}

		b.WriteString(tools.BuildBookingStr(bookings).String())
		b.WriteString(tools.EscapeMarkdownV2("\n"))
	}
	return b.String()
}

func (h *Handler) DailySchedule() {
	h.msgMu.Lock()
	defer h.msgMu.Unlock()
	msg := tgbotapi.NewMessage(h.cfg.GroupChatID, h.buildTodaySchedule())
	msg.ParseMode = "MarkdownV2"

	sent, err := h.bot.Send(msg)
	if err != nil {
		h.log.Error("failed to send DailySchedule", "err", err)
	}
	h.messageID = int64(sent.MessageID)
}

func (h *Handler) wake() {
	if h.messageID == 0 {
		return
	}
	h.msgMu.Lock()
	defer h.msgMu.Unlock()
	edit := tgbotapi.NewEditMessageText(
		h.cfg.GroupChatID,
		int(h.messageID),
		h.buildTodaySchedule(),
	)

	edit.ParseMode = "MarkdownV2"
	// sent, err := h.bot.Send(edit)
	// if err != nil {
	// 	h.log.Error("Wake didnt work out", "err", err)
	// }

	// h.messageID = int64(sent.MessageID)
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("failed to wake", "err", err)
	}
}
