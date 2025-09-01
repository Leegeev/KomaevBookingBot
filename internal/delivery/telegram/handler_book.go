package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

func (h *Handler) handleBook(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /my handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	h.log.Info("Received /book command",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	rooms, err := h.uc.ListRooms(ctx)
	if errors.Is(err, domain.ErrNoRoomsAvailable) {
		h.reply(msg.Chat.ID, getBookNoRoomsAvaibleText())
		return
	} else if err != nil {
		h.log.Error("Failed to list rooms", "user_id", msg.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка:* `%s`", err.Error()))
		h.reply(msg.Chat.ID, "Возникла ошибка при получении доступных комнат.")
		return
	}

	text := "*Выберите переговорку:*"
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(rooms))
	for _, room := range rooms {
		if !room.IsActive {
			continue
		}
		btnText := fmt.Sprintf("#%s", room.Name)
		data := fmt.Sprintf("book:list:%d", room.ID)

		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	if len(rows) == 0 {
		h.reply(msg.Chat.ID, getBookNoRoomsAvaibleText())
		return
	}

	btnText := "Назад"
	data := "book:back"
	btn := tgbotapi.NewInlineKeyboardButtonData(btnText, data)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))

	m := tgbotapi.NewMessage(msg.Chat.ID, EscapeMarkdownV2(text))
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if _, err := h.bot.Send(m); err != nil {
		h.log.Error("Failed to send /my list", "err", err)
	}
}

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

func (h *Handler) handleBookList(ctx context.Context, cq *tgbotapi.CallbackQuery, id int64) {
	h.answerCB(cq, "")
	room, err := h.uc.GetRoom(ctx, id)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: не удалось получить переговорку.")
		return
	}

	// Создаем bookingSession и сохраняем в in-memory storage
	h.sessions.Set(&bookingSession{
		BookState: 1,
		ChatID:    cq.Message.Chat.ID,
		UserID:    cq.From.ID,
		MessageID: cq.Message.MessageID,
		RoomID:    room.ID,
		RoomName:  room.Name,
		Date:      time.Now().In(h.cfg.OfficeTZ).Truncate(24 * time.Hour),
	})

	msg := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		"📅 Выберите дату:",
		buildCalendar(time.Now()),
	)

	msg.ParseMode = "MarkdownV2"
	_, _ = h.bot.Send(msg)
}

func buildCalendar(start time.Time) tgbotapi.InlineKeyboardMarkup {
	// Определим начало недели (понедельник)
	offset := int(start.Weekday()) - 1 // Пн=0 ... Вс=6
	if offset < 0 {
		offset = 6 // если воскресенье
	}
	monday := start.AddDate(0, 0, -offset)

	// Строка 1 — навигация
	row1 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⏪", "book:calendar_nav:-1"),
		tgbotapi.NewInlineKeyboardButtonData("⏩", "book:calendar_nav:+1"),
	)

	// Строка 2 — дни недели
	daysOfWeek := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	row2 := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	for _, day := range daysOfWeek {
		row2 = append(row2, tgbotapi.NewInlineKeyboardButtonData(day, "noop"))
	}

	// Строка 3 — конкретные даты
	row3 := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	for i := 0; i < 7; i++ {
		day := monday.AddDate(0, 0, i)
		display := day.Format("02.01")
		callback := fmt.Sprintf("book:calendar:%s", day.Format("2006-01-02"))
		row3 = append(row3, tgbotapi.NewInlineKeyboardButtonData(display, callback))
	}

	// Строка 4 — Назад
	row4 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🔙 Назад", "book:list_back"),
	)

	return tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, row4)
}

func (h *Handler) handleBookListBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' на списке комнат", "user_id", cq.From.ID)

	// TODO: вернуться к стартовому экрану (например, список действий)
	h.reply(cq.Message.Chat.ID, "Вы вернулись в главное меню.")
}

func (h *Handler) handleBookCalendar(ctx context.Context, cq *tgbotapi.CallbackQuery, dateStr string) {
	h.answerCB(cq, "")

	// Парсим дату
	date, err := time.ParseInLocation("2006-01-02", dateStr, h.cfg.OfficeTZ)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: неправильная дата")
		return
	}

	session := h.sessions.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	session.Date = date

	h.askTimeInput(ctx, cq.Message.Chat.ID, cq.Message.MessageID)
}

func (h *Handler) handleBookCalendarBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.handleBook(ctx, cq.Message) // вернуться к выбору переговорки
}

func (h *Handler) askTimeInput(ctx context.Context, chatID int64, messageID int) {
	text := getBookAskTimeInputText()
	back := tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "book:calendar_back")
	kb := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(back))

	msg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, EscapeMarkdownV2(text), kb)
	msg.ParseMode = "MarkdownV2"
	_, _ = h.bot.Send(msg)
}

func (h *Handler) handleBookTimeBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	session := h.sessions.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}
	h.showCalendar(ctx, cq.Message.Chat.ID, cq.Message.MessageID, session.Date)
}

func (h *Handler) handleBookDuration(ctx context.Context, cq *tgbotapi.CallbackQuery, durStr string) {
	h.answerCB(cq, "")
	session := h.sessions.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	durF, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: неправильная длительность")
		return
	}

	session.Duration = time.Duration(durF * float64(time.Hour))
	session.EndTime = session.StartTime.Add(session.Duration)

	h.askConfirmation(ctx, cq.Message.Chat.ID, cq.Message.MessageID, session)
}

func (h *Handler) handleBookDurationBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.askTimeInput(ctx, cq.Message.Chat.ID, cq.Message.MessageID)
}

func (h *Handler) handleBookConfirm(ctx context.Context, cq *tgbotapi.CallbackQuery, confirm bool) {
	h.answerCB(cq, "")
	session := h.sessions.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	if confirm {
		err := h.uc.CreateBooking(ctx, *session)
		if err != nil {
			h.reply(cq.Message.Chat.ID, "Ошибка при создании брони: "+err.Error())
			return
		}
		h.reply(cq.Message.Chat.ID, "✅ Бронь успешно создана!")
	} else {
		h.reply(cq.Message.Chat.ID, "❌ Бронь отменена.")
	}

	h.sessions.Delete(cq.From.ID)
}

func (h *Handler) handleBookConfirmBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.askDuration(ctx, cq.Message.Chat.ID, cq.Message.MessageID)
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
