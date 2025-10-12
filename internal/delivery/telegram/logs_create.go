package telegram

import (
	"context"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/internal/usecase"
)

// Step 0. Начало флоу создания записи в журнале
func (h *Handler) handleLogCreate0(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogSoglasheniya",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextLogChooseType.String())
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tools.BuildLogCreateKB("create")
	go func() {
		if _, err := h.bot.Send(m); err != nil {
			h.log.Error("Failed to handle /handleLogCreate0 on rooms list", "err", err)
		}
	}()
}

// Step 1_0 (тип выбран)
// Хендлер выбора типа записи и переход к выбору даты
func (h *Handler) handleLogCreate1_0(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	logType := parts[2]

	// Создаем bookingSession и сохраняем в in-memory storage
	h.logSession.Set(&tools.LogsSession{
		State:     tools.StateProcessingLogCreating,
		Type:      logType,
		UserID:    cq.From.ID,
		MessageID: cq.Message.MessageID,
	})

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

// Step 1_1 (календарь сдвинут)
// Хендлер сдвига календаря
func (h *Handler) handleLogСreate1_1(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	// 1. Парсим направление навигации
	parts := strings.Split(cq.Data, ":")
	shift, _ := strconv.ParseInt(parts[2], 10, 64) // -3 -2 -1 1 2 3

	// Если пользователь хочет передвинуться в прошлое, игнорируем
	if shift < -1 || shift > 0 {
		return
	}
	// 2. Обновляем только клавиатуру
	editMarkup := tgbotapi.NewEditMessageReplyMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildLogCalendarKB(shift),
	)
	go func() {
		if _, err := h.bot.Request(editMarkup); err != nil {
			h.log.Error("failed to edit handleLogСreate1_1 calendar inline keyboard", "err", err)
		}
	}()
}

// Step 2 (дата выбрана)
// Парсер выбора даты и переход к вводу времени
func (h *Handler) handleLogCreate2(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	dateStr := parts[2] // формат даты предполагается как "YYYY-MM-DD"

	// Парсим дату
	date, err := time.ParseInLocation("2006-01-02", dateStr, h.cfg.OfficeTZ)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: неправильная дата")
		return
	}

	session := h.logSession.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	session.Date = date
	session.State = tools.StateInputingName // ⚠️⚠️⚠️ сделать проверку на ввод

	edit := tgbotapi.EditMessageTextConfig{}

	// Запрашивается, либо ввод ФИО
	// либо ввод ДОВЕРИТЕЛЯ
	if h.logsUC.GetUser(ctx, cq.From.ID) != nil {
		// TODO: если пользователь не найден, запросить ФИО
		// Тогда сразу к следующему шагу
	} else {
		edit = tgbotapi.NewEditMessageTextAndMarkup(
			cq.Message.Chat.ID,
			cq.Message.MessageID,
			tools.TextLogAskName.String(),
			tgbotapi.NewInlineKeyboardMarkup(
				[]tgbotapi.InlineKeyboardButton{
					tools.BuildBackInlineKBButton("log:step2_back"),
				}),
		)
	}

	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on handleLogCreate2", "err", err)
		}
	}()
}

// Step 3 ФИО введено
// Парсер Фио и ввод доверителя
func (h *Handler) handleLogCreate3(ctx context.Context, msg *tgbotapi.Message) {
	FIO := strings.TrimSpace(msg.Text)
	session := h.logSession.Get(msg.From.ID)
	if session == nil {
		h.reply(msg.Chat.ID, "Сессия не найдена")
		return
	}

	// Сохраняем ФИО в БД пользователей
	if err := h.logsUC.CreateUser(ctx, msg.From.ID, FIO); err != nil {
		h.log.Error("Failed to save user FIO", "err", err)
		h.notifyAdmin("Ошибка при сохранении ФИО")
		h.reply(msg.Chat.ID, "Ошибка при сохранении ФИО. Попробуйте еще раз.")
		return
	}

	session.State = tools.StateInputingDoveritel // следующий шаг - ввод доверителя
	session.UserName = FIO

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		msg.Chat.ID,
		session.MessageID,
		tools.TextLogAskName.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on handleLogCreate3", "err", err)
	}

	newMsg := tgbotapi.NewMessage(msg.Chat.ID, tools.TextLogAskDoveritel.String())
	newMsg.ParseMode = "MarkdownV2"

	replyKB := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tools.BuildBackInlineKBButton("log:step3_back"),
		},
	)
	newMsg.ReplyMarkup = replyKB

	go func() {
		sentMsg, err := h.bot.Send(newMsg)
		if err != nil {
			h.log.Error("Failed to send a new message on handleLogCreate3", "err", err)
			return
		}
		// обновляем messageID в сессии
		// чтобы при НАЗАД можно было его отредактировать
		// (иначе будет редактироваться первое сообщение, а не текущее)
		session.MessageID = sentMsg.MessageID
	}()
}

// Step 4 Доверитель введен. Сразу сюда, если ФИО есть в БД
// Парсер Доверителя и Ввод комментария
func (h *Handler) handleLogCreate4(ctx context.Context, msg *tgbotapi.Message) {
	Doveritel := strings.TrimSpace(msg.Text)
	session := h.logSession.Get(msg.From.ID)
	if session == nil {
		h.reply(msg.Chat.ID, "Сессия не найдена")
		return
	}

	session.State = tools.StateInputingDoveritel
	session.Doveritel = Doveritel

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		msg.Chat.ID,
		session.MessageID,
		tools.TextLogAskDoveritel.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on handleLogCreate3", "err", err)
	}

	newMsg := tgbotapi.NewMessage(msg.Chat.ID, tools.TextLogAskComment.String())
	newMsg.ParseMode = "MarkdownV2"

	replyKB := tgbotapi.NewInlineKeyboardMarkup(
		[]tgbotapi.InlineKeyboardButton{
			tools.BuildBackInlineKBButton("log:step4_back"),
		},
	)
	newMsg.ReplyMarkup = replyKB

	go func() {
		sentMsg, err := h.bot.Send(newMsg)
		if err != nil {
			h.log.Error("Failed to send a new message on handleLogCreate4", "err", err)
			return
		}
		// обновляем messageID в сессии
		// чтобы при НАЗАД можно было его отредактировать
		// (иначе будет редактироваться первое сообщение, а не текущее)
		session.MessageID = sentMsg.MessageID
	}()
}

// Step 5 Комментарий введен.
// Парсер комментария и Подтверждение создания
func (h *Handler) handleLogCreate5(ctx context.Context, msg *tgbotapi.Message) {
	comment := strings.TrimSpace(msg.Text)
	session := h.logSession.Get(msg.From.ID)
	if session == nil {
		h.reply(msg.Chat.ID, "Сессия не найдена")
		return
	}

	session.State = tools.StateCreateConfirming
	session.Comment = comment

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		msg.Chat.ID,
		session.MessageID,
		tools.BuildLogConfirmationStr(session).String(),
		tools.BuildConfirmationKB("log"),
	)

	edit.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(edit); err != nil {
			h.log.Error("Failed to edit message on handleLogCreate5", "err", err)
		}
	}()
}

// Step 6 Финиш
// Парсер подтверждения и Создание записи
func (h *Handler) handleLogCreate6(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	confirm, _ := strconv.ParseInt(parts[2], 10, 64) // 0 || 1

	session := h.logSession.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	cmd := usecase.CreateLogCmd{
		UserID:    domain.UserID(session.UserID),
		UserName:  session.UserName,
		Type:      session.Type,
		Date:      session.Date,
		Doveritel: session.Doveritel,
		Comment:   session.Comment,
	}

	edit := tgbotapi.NewEditMessageReplyMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildBlankInlineKB(),
	)
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to EDIT message on handleLogCreate6 confirmation", "err", err)
	}

	var replyText string
	if confirm == 1 {
		num, err := h.logsUC.CreateLog(ctx, cmd)
		if err != nil {
			// TODO: обработать ошибку
		}
		replyText = tools.BuildLogConfirmedStr(num).String()
	} else {
		replyText = tools.TextLogNo.String()
	}
	h.sessions.Delete(cq.From.ID)

	role, err := h.getRole(cq.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", cq.From.ID, "username", cq.From.UserName)
		role = tools.Member
	}
	newMsg := tgbotapi.NewMessage(cq.Message.Chat.ID, replyText)
	newMsg.ReplyMarkup = tools.BuildLogMainKB(role)
	newMsg.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(newMsg); err != nil {
			h.log.Error("Failed to SEND a new message on handleLogCreate6 confirmation", "err", err)
		}
	}()
}
