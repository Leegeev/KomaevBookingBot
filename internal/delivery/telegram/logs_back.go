package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

func (h *Handler) handleLogCreateBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Debug("User clicked назад на handleLogCreate0. обработчик: handleLogCreateBack") // это выбор типа

	edit := tgbotapi.NewEditMessageReplyMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildBlankInlineKB(),
	)
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("handleMyBack failed to hide inline KB", "err", err)
	}

	role, err := h.getRole(cq.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", cq.From.ID, "username", cq.From.UserName)
		role = tools.Member
	}
	msg := tgbotapi.NewMessage(cq.Message.Chat.ID, "Выберите действие:")
	msg.ReplyMarkup = tools.BuildLogMainKB(role)
	msg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Error("failed to send reply keyboard", "err", err)
	}
}

func (h *Handler) handleLogCalendarBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Debug("User clicked назад на handleLogCreate1_0. обработчик: handleLogCalendarBack") // это календарь

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextLogChooseType.String(),
		tools.BuildLogCreateKB("create"),
	)
	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on handleLogCreate1_0 list", "err", err)
		}
	}()
}

func (h *Handler) handleLogStep2Back(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Debug("User clicked назад на handleLogCreate2. обработчик: handleLogStep2Back") // это он тыкнул дату и должен вводить имя, но мы вернем его на календарь
	session := h.logSession.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}
	session.State = tools.StateProcessingLogCreating

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextLogCalendar.String(),
		tools.BuildLogCalendarKB(0),
	)
	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on handleLogCreate1_0 list", "err", err)
		}
	}()
}

func (h *Handler) handleLogStep3Back(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Debug("User clicked назад на handleLogCreate3. обработчик: handleLogStep3Back") // это пользователь уже на вводе доверителя

	session := h.logSession.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	if session.Registration { // Если человек до этого вводил ФИО, то вернуть на ввод ФИО
		session.State = tools.StateInputingName
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			cq.Message.Chat.ID,
			cq.Message.MessageID,
			tools.TextLogAskName.String(),
			tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{
					tools.BuildBackInlineKBButton("log:step2_back"),
				}),
		)
		edit.ParseMode = "MarkdownV2"
		go func() {
			if _, err := h.bot.Send(edit); err != nil {
				h.log.Error("Failed to edit message on handleLogCreate2", "err", err)
			}
		}()
		return
	} // Иначе вернуть на выбор даты

	session.State = tools.StateProcessingLogCreating
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextLogCalendar.String(),
		tools.BuildLogCalendarKB(0),
	)
	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on handleLogCreate1_0 list", "err", err)
		}
	}()
}

func (h *Handler) handleLogStep4Back(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Debug("User clicked назад на handleLogCreate4. обработчик: handleLogStep4Back") // это пользователь уже на вводе комментария
	session := h.logSession.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}
	session.State = tools.StateInputingDoveritel

	replyKB := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tools.BuildBackInlineKBButton("log:step3_back"),
		},
	)

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextLogAskDoveritel.String(),
		replyKB,
	)
	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on handleLogCreate1_0 list", "err", err)
		}
	}()
}

func (h *Handler) handleLogConfirmBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Debug("User clicked назад на handleLogCreate5. обработчик: handleLogConfirmBack") // Парсер комментария и Подтверждение создания
	session := h.logSession.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}
	session.State = tools.StageInputingComment

	replyKB := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tools.BuildBackInlineKBButton("log:step4_back"),
		},
	)

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextLogAskComment.String(),
		replyKB,
	)
	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on handleLogCreate1_0 list", "err", err)
		}
	}()
}
