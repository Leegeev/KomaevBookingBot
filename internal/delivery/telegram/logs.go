package telegram

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
	filePath, err := h.logsUC.CreateExcelReport(ctx)
	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка при создании отчета 😔")
		h.log.Error("CreateExcelReport error", "err", err)
		return
	}
	defer os.Remove(filePath)

	// Отправляем файл пользователю
	doc := tgbotapi.NewDocument(msg.Chat.ID, tgbotapi.FilePath(filePath))
	doc.Caption = "📊 Отчёт по Журналам"
	if _, err := h.bot.Send(doc); err != nil {
		h.log.Error("Failed to send Excel file", "err", err)
		h.reply(msg.Chat.ID, "Не удалось отправить файл 😔")
		h.notifyAdmin("Failed to send Excel file to user")
		return
	}
}

func (h *Handler) handleLogFind(ctx context.Context, msg *tgbotapi.Message) {
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

	// Проверяем длину и префикс
	if len(num) < 3 {
		h.reply(msg.Chat.ID, "Некорректный номер. Пример: ЭС12345 или ЭЗ12345")
		return
	}

	// Определяем тип и отделяем числовую часть
	var (
		prefix string
		idStr  string
	)

	switch {
	case strings.HasPrefix(num, "ЭС"):
		prefix = "ЭС"
		idStr = strings.TrimPrefix(num, "ЭС")
	case strings.HasPrefix(num, "ЭЗ"):
		prefix = "ЭЗ"
		idStr = strings.TrimPrefix(num, "ЭЗ")
	default:
		h.reply(msg.Chat.ID, "Введите корректный номер: ЭС12345 или ЭЗ12345")
		return
	}

	// Преобразуем числовую часть в int64
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.reply(msg.Chat.ID, "Некорректный формат номера. Пример: ЭС12345 или ЭЗ12345")
		return
	}

	// Обработка в зависимости от типа
	switch prefix {
	case "ЭС":
		record, err := h.logsUC.GetSoglasheniyaById(ctx, id)
		// if err != nil {
		// 	h.reply(msg.Chat.ID, "Ошибка при поиске соглашения")
		// 	h.log.Error("GetSoglasheniyaById error", "err", err)
		// 	h.notifyAdmin("Не получилось обработать LogFind из-за ошибки в UC")
		// 	return
		// }
		// h.reply(msg.Chat.ID, fmt.Sprintf("Найден номер соглашения: ЭС%d", id))

	case "ЭЗ":
		record, err := h.logsUC.GetZaprosById(ctx, id)
		// if err != nil {
		// 	h.reply(msg.Chat.ID, "Ошибка при поиске запроса")
		// 	h.log.Error("GetZaprosById error", "err", err)
		// 	h.notifyAdmin("Не получилось обработать LogFind из-за ошибки в UC")
		// 	return
		// }
		// h.reply(msg.Chat.ID, fmt.Sprintf("Найден номер запроса: ЭЗ%d", id))
	}

	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка при поиске соглашения")
		h.log.Error("GetSoglasheniyaById error", "err", err)
		h.notifyAdmin("Не получилось обработать LogFind из-за ошибки в UC")
		return
	}
	// Обработать найденную запись. Вывести сведения
	h.reply(msg.Chat.ID, fmt.Sprintf("Найден номер запроса: ЭЗ%d", id))
}
