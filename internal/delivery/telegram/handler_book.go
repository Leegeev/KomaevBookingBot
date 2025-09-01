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
	h.sessions.Set(cq.From.ID, &bookingSession{
		BookState: 1,
		ChatID:    cq.Message.Chat.ID,
		UserID:    int64(cq.From.ID),
		MessageID: cq.Message.MessageID,
		RoomID:    strconv.FormatInt(room.ID, 10),
		RoomName:  room.Name,
		Date:      time.Now().In(h.cfg.OfficeTZ).Truncate(24 * time.Hour),
	})

	h.showCalendar(ctx, cq.Message.Chat.ID, cq.Message.MessageID, time.Now().In(h.cfg.OfficeTZ))
}

func (h *Handler) handleBookListBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' на списке комнат", "user_id", cq.From.ID)

	// TODO: вернуться к стартовому экрану (например, список действий)
	h.reply(cq.Message.Chat.ID, "Вы вернулись в главное меню.")
}

func (h *Handler) handleBookCalendar(ctx context.Context, cq *tgbotapi.CallbackQuery, dateStr string) {
	/*

		тут предполагается, что ты ждёшь текстовое сообщение с временем начала
		→
		обрабатывай его в handleMessage и сверяй BookState.

	*/
	h.answerCB(cq, "")
	h.log.Info("Date selected", "user_id", cq.From.ID, "date", dateStr)

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Некорректная дата")
		return
	}

	session := h.getSession(cq.From.ID)
	session.Date = date
	h.saveSession(session)

	// Переход к выбору времени
	h.reply(cq.Message.Chat.ID, fmt.Sprintf("Вы выбрали дату *%s*. Теперь введите время начала (например, 13:30).", date.Format("02.01")))
}

func (h *Handler) handleBookCalendarBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' на календаре", "user_id", cq.From.ID)

	// TODO: вернуться к списку комнат
	h.reply(cq.Message.Chat.ID, "Назад к выбору комнаты.")
}

func (h *Handler) handleBookTimeBack(ctx context.Context, cq *tgbotapi.CallbackQuery, duration string) {
}

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

func (h *Handler) handleBookConfirm(ctx context.Context, cq *tgbotapi.CallbackQuery, confirmed bool) {
	h.answerCB(cq, "")
	if confirmed {
		h.log.Info("User confirmed booking", "user_id", cq.From.ID)
		// TODO: сохранить бронь в БД
		h.reply(cq.Message.Chat.ID, "✅ Ваша бронь успешно создана.")
	} else {
		h.log.Info("User cancelled booking confirmation", "user_id", cq.From.ID)
		// TODO: вернуться к выбору параметров
		h.reply(cq.Message.Chat.ID, "Бронирование отменено.")
	}
}

func (h *Handler) handleBookConfirmBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' на подтверждении", "user_id", cq.From.ID)

	// TODO: вернуться к выбору длительности
	h.reply(cq.Message.Chat.ID, "Назад к выбору длительности.")
}
