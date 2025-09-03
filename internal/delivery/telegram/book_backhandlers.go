package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Step 0.
// Хендлер кнопки назад
func (h *Handler) handleBookListBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' на списке комнат", "user_id", cq.From.ID)

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		EscapeMarkdownV2(mainMenuText.String()),
		blankInlineKB(),
	)

	msg.ParseMode = "MarkdownV2"
	_, _ = h.bot.Send(msg)

	// TODO: вернуться к стартовому экрану (например, список действий)
	// h.reply(cq.Message.Chat.ID, "Вы вернулись в главное меню.")
}

// Step 1.
// Хендлер кнопки назад в календаре
func (h *Handler) handleBookCalendarBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "") // Убираем часики и не показываем уведомление

	h.log.Info("User clicked 'Назад' on calendar", "user_id", cq.From.ID)

	rows, err := h.buildRoomListKB(ctx, cq.From.ID)
	if err != nil {
		h.log.Error("Failed to build room list on calendar back", "user_id", cq.From.ID, "err", err)
		// Не можем редактировать сообщение, поэтому отправим новое
		h.reply(cq.Message.Chat.ID, err.Error())
		return
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		EscapeMarkdownV2(bookIntroductionText.String()),
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)
	edit.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on calendar back", "err", err)
	}
}

func (h *Handler) handleBookTimepickBack(ctx context.Context, msg *tgbotapi.Message) {
	return
}

func (h *Handler) handleBookDurationBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.askTimeInput(ctx, cq.Message.Chat.ID, cq.Message.MessageID)
}

func (h *Handler) handleBookConfirmBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.askDuration(ctx, cq.Message.Chat.ID, cq.Message.MessageID)
}
