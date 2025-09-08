package telegram

import (
	"context"
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

// Step 0.
// Хендлер кнопки назад
func (h *Handler) handleBookListBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' на списке комнат", "user_id", cq.From.ID)

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextMainMenu.String(),
		tools.BuildBlankInlineKB(),
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

	rooms, err := h.uc.ListRooms(ctx)
	if errors.Is(err, domain.ErrNoRoomsAvailable) {
		h.reply(cq.From.ID, tools.TextBookNoRoomsAvailable.String())
		return
	} else if err != nil {
		h.log.Error("Failed to list rooms", "user_id", cq.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /book:* `%s`", err.Error()))
		h.reply(cq.From.ID, tools.TextBookNoRoomsErr.String())
		return
	}

	rows := tools.BuildRoomListKB(ctx, rooms)
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextBookIntroduction.String(),
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)
	edit.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on calendar back", "err", err)
	}
}

func (h *Handler) handleBookTimepickBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
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
