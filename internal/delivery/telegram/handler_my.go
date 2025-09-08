package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

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

		data := fmt.Sprintf("my:list:%d", bk.ID)

		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	btnText := "Назад"
	data := "my:back"
	btn := tgbotapi.NewInlineKeyboardButtonData(btnText, data)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))

	m := tgbotapi.NewMessage(msg.Chat.ID, tools.EscapeMarkdownV2(text))
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if _, err := h.bot.Send(m); err != nil {
		h.log.Error("Failed to send /my list", "err", err)
	}
}

// func (h *Handler) handleMyCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {}

// этот обработчик вызывается из какого-то роутера.
func (h *Handler) handleMyCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	// data формата: my:list:<id> | my:cancel:<id> | my:list_back | my:cancel_back

	parts := strings.Split(cq.Data, ":")
	if len(parts) < 2 || parts[0] != "my" {
		return
	}
	action := parts[1]

	switch action {
	case "list":
		if len(parts) != 3 {
			return
		}
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return
		}
		// Грузим бронь и показываем подтверждение
		h.handleMyListCallback(ctx, cq, id)

	case "cancel":
		if len(parts) != 3 {
			return
		}
		id, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return
		}
		h.handleMyCancelCallback(ctx, cq, id)

		/*
			// Отменяем бронь
			if err := h.uc.CancelBooking(ctx, id); err != nil {
				h.answerCB(cq, "Не удалось отменить бронь 😕")
				return
			}
			// Сообщение-подтверждение
			text := "✅ Бронь отменена."
			edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID, EscapeMarkdownV2(text))
			edit.ParseMode = "MarkdownV2"
			// Уберём клавиатуру
			editReply := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tgbotapi.InlineKeyboardMarkup{})
			if _, err := h.bot.Send(edit); err != nil {
				h.log.Error("Failed to edit message to 'canceled'", "err", err)
			}
			if _, err := h.bot.Send(editReply); err != nil {
				h.log.Error("Failed to clear keyboard", "err", err)
			}
			h.answerCB(cq, "Готово")
		*/

	case "list_back":
		// клавиатура с главного меню.

	case "cancel_back":
		// вывести результат работы /my

	}
}

func (h *Handler) handleMyCancelCallback(ctx context.Context, cq *tgbotapi.CallbackQuery, id BookingID) {
	return
}

func (h *Handler) handleMyListCallback(ctx context.Context, cq *tgbotapi.CallbackQuery, id BookingID) {
	return
}
