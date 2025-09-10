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
		h.reply(msg.Chat.ID, "У вас нет броней.")
		return
	}

	kb := tools.BuildMyListKB(bookings, h.cfg.OfficeTZ)

	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextMyIntroduction.String())
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = kb

	if _, err := h.bot.Send(m); err != nil {
		h.log.Error("Failed to send /my list", "err", err)
	}
}

// TODO: вернуться к стартовому экрану (например, список действий)
// h.reply(cq.Message.Chat.ID, "Вы вернулись в главное меню.")
func (h *Handler) handleMyBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("handleMyBack", "user_id", cq.From.ID)

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextRedirectingToMainMenu.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("handleMyBack failed to hide inline KB", "err", err)
	}

	role, _ := h.getRole(ctx, cq.From.ID)
	replyKB := tools.BuildMainMenuKB(role)

	msg := tgbotapi.NewMessage(cq.Message.Chat.ID, "Главное меню:")
	msg.ReplyMarkup = replyKB

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Error("failed to send main menu", "err", err)
	}
}

func (h *Handler) handleMyList(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	// my:reschedule:<id>
	// my:cancel:<id>
	// my:list_back

	// [перенести] [отменить]
	// [назад]
	h.answerCB(cq, "")
	h.log.Info("handleMyList", "data", cq.Data, "user", cq.From.UserName)

	parts := strings.Split(cq.Data, ":")
	id, _ := strconv.ParseInt(parts[2], 10, 64) // id of the picked booking.

	// Грузим бронь
	bk, err := h.uc.GetById(ctx, id)
	if err != nil {
		h.log.Error("Failed to get booking for my:list", "err", err, "user_id", cq.From.ID, "bk_id", id)
		h.answerCB(cq, "Не удалось загрузить бронь 😕")
		return
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildMyOperationStr(bk),
		tools.BuildMyOperationsKB(id),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on book list", "err", err)
	}
}

// func (h *Handler) handleMyReschedule(ctx context.Context, cq *tgbotapi.CallbackQuery) {
// 	h.answerCB(cq, "")
// 	h.log.Info("handleMyReschedule", "data", cq.Data, "user", cq.From.UserName)

// 	parts := strings.Split(cq.Data, ":")
// 	id, _ := strconv.ParseInt(parts[2], 10, 64) // id of the picked booking.
// 	// delete current booking
// 	// start booking new flow

// }

func (h *Handler) handleMyCancel(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("handleMyReschedule", "data", cq.Data, "user", cq.From.UserName)

	parts := strings.Split(cq.Data, ":")
	id, _ := strconv.ParseInt(parts[2], 10, 64) // id of the picked booking.

	// bk, err := h.uc.GetById(ctx, id)
	// if err != nil {
	// 	h.log.Error("Failed to get booking for my:cancel", "err", err, "user_id", cq.From.ID, "bk_id", id)
	// 	h.answerCB(cq, "Не удалось загрузить бронь 😕")
	// 	return
	// }

	// if cq.Message.From.ID == int64(bk.CreatedBy) {
	// 	h.log.Info("User owns the booking", "userID", cq.From.ID, "bookingID", id)
	// 	h.uc.CancelBooking(ctx, id)
	// }

	if err := h.uc.CancelBooking(ctx, id); err != nil {
		h.log.Error("Failed to cansel booking", "user_id", cq.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /my_cancel:* `%s`", err.Error()))
		h.reply(cq.From.ID, tools.TextMyBookingCancelErr.String())
		return
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextMyBookingCancelled.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on book list", "err", err)
	}
}

func (h *Handler) handleMyListBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("handleMyListBack", "data", cq.Data, "user", cq.From.UserName)

	bookings, err := h.uc.ListUserBookings(ctx, int64(cq.From.ID))
	if err != nil {
		h.log.Error("Failed to list user bookings", "user_id", cq.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка:* `%s`", err.Error()))
		h.reply(cq.Message.Chat.ID, "Возникла ошибка при получении ваших броней.")
		return
	}

	if len(bookings) == 0 {
		h.reply(cq.Message.Chat.ID, "У вас нет броней.")
		return
	}

	kb := tools.BuildMyListKB(bookings, h.cfg.OfficeTZ)

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextBookIntroduction.String(),
		kb,
	)
	edit.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on calendar back", "err", err)
	}
}

/*


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


			// // Отменяем бронь
			// if err := h.uc.CancelBooking(ctx, id); err != nil {
			// 	h.answerCB(cq, "Не удалось отменить бронь 😕")
			// 	return
			// }
			// // Сообщение-подтверждение
			// text := "✅ Бронь отменена."
			// edit := tgbotapi.NewEditMessageText(cq.Message.Chat.ID, cq.Message.MessageID, EscapeMarkdownV2(text))
			// edit.ParseMode = "MarkdownV2"
			// // Уберём клавиатуру
			// editReply := tgbotapi.NewEditMessageReplyMarkup(cq.Message.Chat.ID, cq.Message.MessageID, tgbotapi.InlineKeyboardMarkup{})
			// if _, err := h.bot.Send(edit); err != nil {
			// 	h.log.Error("Failed to edit message to 'canceled'", "err", err)
			// }
			// if _, err := h.bot.Send(editReply); err != nil {
			// 	h.log.Error("Failed to clear keyboard", "err", err)
			// }
			// h.answerCB(cq, "Готово")


	case "list_back":
		// клавиатура с главного меню.

	case "cancel_back":
		// вывести результат работы /my

	}
}

*/
