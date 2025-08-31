package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func (h *Handler) handleMyCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {

}
