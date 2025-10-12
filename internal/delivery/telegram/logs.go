package telegram

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

func (h *Handler) handleLog(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLog",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	role, err := h.getRole(msg.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", msg.From.ID, "username", msg.From.UserName)
		role = tools.Member
	}

	kb := tools.BuildLogMainKB(role)
	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextLogMainMenu.String())
	m.ReplyMarkup = kb
	m.ParseMode = "MarkdownV2"

	go func() {
		if _, err := h.bot.Send(m); err != nil {
			h.log.Error("Failed to send a new message on handleLog", "err", err)
		}
	}()
}

func (h *Handler) handleLogMy0(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogZaprosi",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextLogChooseType.String())
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tools.BuildLogCreateKB("my")
	go func() {
		if _, err := h.bot.Send(m); err != nil {
			h.log.Error("Failed to handle /handleLogMy0 on rooms list", "err", err)
		}
	}()
}

func (h *Handler) handleLogMy1(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	logType := parts[2]
	msgText := ""
	if logType == "sogl" {
		logs := h.logsUC.GetSoglasheniyaByUserID(ctx, cq.From.ID)
		msgText = tools.BuildLogListStr(logs)
	}

	if logType == "zapros" {
		logs := h.logsUC.GetZaprosiByUserId(ctx, cq.From.ID)
		msgText = tools.BuildLogListStr(logs)
	}

	msg := tgbotapi.NewMessage(
		cq.Message.Chat.ID,
		msgText,
	)

	msg.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(msg); err != nil {
			h.log.Error("Failed to edit message on handleLogMy1 list", "err", err)
		}
	}()
}

func (h *Handler) handleLogExport(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogCreate",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
	// TODO
	h.reply(msg.Chat.ID, "Экспорт журнала пока не реализован.")
}

func (h *Handler) handleLogFind(ctx context.Context, msg *tgbotapi.Message) {
	// TODO
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogFind",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	role, err := h.getRole(msg.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", msg.From.ID, "username", msg.From.UserName)
		role = tools.Member
	}

	if !tools.CheckRoleIsAdmin(role) {
		msgText := "Команда доступна только администраторам."
		h.reply(msg.Chat.ID, msgText)
		return
	}

	num := strings.TrimSpace(msg.Text)

	if strings.HasPrefix(num, "ЭС") {
		// Проверка корректности номера соглашения (ЭС)
		if len(num) < 3 {
			h.reply(msg.Chat.ID, "Некорректный номер соглашения. Пример: ЭС12345")
			return
		}
		// Здесь будет обработка соглашения
		h.reply(msg.Chat.ID, "Найден номер соглашения: "+num)
		return
	}

	if strings.HasPrefix(num, "ЭЗ") {
		// Проверка корректности номера запроса (ЭЗ)
		if len(num) < 3 {
			h.reply(msg.Chat.ID, "Некорректный номер запроса. Пример: ЭЗ12345")
			return
		}
		// Здесь будет обработка запроса
		h.reply(msg.Chat.ID, "Найден номер запроса: "+num)
		return
	}

	h.reply(msg.Chat.ID, "Пожалуйста, введите корректный номер соглашения (ЭС...) или запроса (ЭЗ...)")
}
