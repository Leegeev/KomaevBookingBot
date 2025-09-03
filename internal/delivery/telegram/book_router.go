package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Роутер /book.
// Обрабатывает все callback'и
func (h *Handler) handleBookCallback(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	parts := strings.Split(cq.Data, ":")
	if len(parts) < 2 || parts[0] != "book" {
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
		h.handleBookList(ctx, cq, id)

	case "list_back":
		h.handleBookListBack(ctx, cq)

	case "calendar":
		if len(parts) != 3 {
			return
		}
		date := parts[2] // формат даты предполагается как "YYYY-MM-DD"
		h.handleBookCalendar(ctx, cq, date)

	case "calendar_back":
		h.handleBookCalendarBack(ctx, cq)

	case "timepick_back":
		h.handleBookTimeBack(ctx, cq)

	case "duration":
		if len(parts) != 3 {
			return
		}
		duration := parts[2] // строка вроде "0.5", "1.0", "2.5"
		h.handleBookDuration(ctx, cq, duration)

	case "duration_back":
		h.handleBookDurationBack(ctx, cq)

	case "confirm":
		if len(parts) != 3 {
			return
		}
		val := parts[2]
		confirmed := val == "true"
		h.handleBookConfirm(ctx, cq, confirmed)

	case "confirm_back":
		h.handleBookConfirmBack(ctx, cq)

	default:
		h.log.Warn("Unknown book callback", "data", cq.Data)
	}
}

/*

func (h *Handler) handleBookDuration(ctx context.Context, cq *tgbotapi.CallbackQuery, durationStr string) {
	h.answerCB(cq, "")
	h.log.Info("Duration selected", "user_id", cq.From.ID, "duration", durationStr)

	dur, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Некорректный формат длительности")
		return
	}

	session := h.getSession(cq.From.ID)
	session.Duration = time.Duration(dur * float64(time.Hour))
	session.EndTime = session.StartTime.Add(session.Duration)
	h.saveSession(session)

	text := fmt.Sprintf(
		"*Подтвердите бронь:*\n\n"+
			"• Комната: *%s*\n"+
			"• Дата: *%s*\n"+
			"• Время: *%s–%s*\n",
		session.RoomName,
		session.Date.Format("02.01"),
		session.StartTime.Format("15:04"),
		session.EndTime.Format("15:04"),
	)

	// Кнопки "Да" / "Нет"
	yes := tgbotapi.NewInlineKeyboardButtonData("✅ Да", "book:confirm:true")
	no := tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "book:confirm:false")
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(yes, no))

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		session.ChatID,
		session.MessageID,
		EscapeMarkdownV2(text),
		kb,
	)
	msg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Error("Failed to send confirm screen", "err", err)
	}
}

func (h *Handler) handleBookDurationBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' на длительности", "user_id", cq.From.ID)

	// TODO: вернуться к выбору времени
	h.reply(cq.Message.Chat.ID, "Назад к выбору времени.")
}

*/
