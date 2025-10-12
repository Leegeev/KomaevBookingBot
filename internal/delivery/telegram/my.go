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

	bookings, err := h.uc.ListUserBookings(ctx, int64(msg.From.ID))
	if err != nil {
		h.log.Error("Failed to list user bookings", "user_id", msg.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка:* `%s`", err.Error()))
		h.reply(msg.Chat.ID, "Возникла ошибка при получении ваших броней. Тех. поддержка уже уведомлена.")
		return
	}

	if len(bookings) == 0 {
		h.reply(msg.Chat.ID, "У вас нет броней.")
		return
	}

	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextMyIntroduction.String())
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	m.ReplyMarkup = tools.BuildMyListKB(bookings, h.cfg.OfficeTZ)
	go func() {
		if _, err := h.bot.Send(m); err != nil {
			h.log.Error("Failed to send /my list", "err", err)
		}
	}()
}

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

	role, err := h.getRole(cq.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", cq.From.ID, "username", cq.From.UserName)
		role = tools.Member
	}
	replyKB := tools.BuildMainMenuKB(role)

	msg := tgbotapi.NewMessage(cq.Message.Chat.ID, tools.TextMainMenu.String())
	msg.ReplyMarkup = replyKB
	go func() {
		if _, err := h.bot.Send(msg); err != nil {
			h.log.Error("failed to send main menu", "err", err)
		}
	}()
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
		tools.BuildMyOperationStr(bk).String(),
		tools.BuildMyOperationsKB(id),
	)

	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on book list", "err", err)
		}
	}()
}

func (h *Handler) handleMyCancel(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("handleMyCancel", "data", cq.Data, "user", cq.From.UserName)

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

	go h.wake()

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextMyBookingCancelled.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on book list", "err", err)
		}
	}()
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
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on calendar back", "err", err)
		}
	}()
}
