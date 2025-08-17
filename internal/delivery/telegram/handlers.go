package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/internal/usecase"
)

type bookingSession struct {
	RoomID   int64
	DayLocal time.Time // полночь локального дня (OfficeTZ)

	StartH, StartM *int
	EndH, EndM     *int
}

/* ---------- /start /help ---------- */

func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	text := "Привет! Я бот бронирования переговорок.\n\n" +
		"/book — забронировать\n" +
		"/my — мои брони\n" +
		"/rooms — список переговорок\n" +
		"/cancel — отменить бронь\n" +
		"/help — справка\n" +
		"/create_room — создать переговорку (админ)\n" +
		"/deactivate_room — деактивировать переговорку (админ)"
	h.reply(msg.Chat.ID, text)
}

func (h *Handler) handleHelp(ctx context.Context, msg *tgbotapi.Message) {
	text := "*Команды:*\n" +
		"• /book — выбрать переговорку → день → время начала/окончания кнопками\n" +
		"• /my — показать ваши будущие брони (кнопки «Отменить»)\n" +
		"• /rooms — показать активные переговорки\n" +
		"• /cancel `<id>` — отменить по номеру\n" +
		"• /create_room `<name>` — создать переговорку (админ)\n" +
		"• /deactivate_room `<room_id>` — выключить переговорку (админ)"
	h.reply(msg.Chat.ID, text)
}

/* ---------- /rooms ---------- */

func (h *Handler) handleRooms(ctx context.Context, msg *tgbotapi.Message) {
	rooms, err := h.uc.ListRooms(ctx)
	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	if len(rooms) == 0 {
		h.reply(msg.Chat.ID, "Активных переговорок нет.")
		return
	}
	var b strings.Builder
	b.WriteString("*Переговорки:*\n")
	for _, r := range rooms {
		fmt.Fprintf(&b, "• %d — %s\n", r.ID, r.Name)
	}
	h.reply(msg.Chat.ID, b.String())
}

/* ---------- /my ---------- */

func (h *Handler) handleMy(ctx context.Context, msg *tgbotapi.Message) {
	bookings, err := h.uc.ListUserBookings(ctx, int64(msg.From.ID))
	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	if len(bookings) == 0 {
		h.reply(msg.Chat.ID, "У вас нет будущих броней.")
		return
	}

	kbRows := [][]tgbotapi.InlineKeyboardButton{}
	var b strings.Builder
	b.WriteString("*Ваши брони:*\n")
	for _, bk := range bookings {
		start := bk.Range.Start.In(h.cfg.OfficeTZ) // если у тебя bk.Interval — замени
		end := bk.Range.End.In(h.cfg.OfficeTZ)
		fmt.Fprintf(&b, "• #%d — %s %02d:%02d–%02d:%02d\n",
			bk.ID, start.Format("2006-01-02"), start.Hour(), start.Minute(), end.Hour(), end.Minute())

		cb := tgbotapi.NewInlineKeyboardButtonData("Отменить", fmt.Sprintf("c:%d", bk.ID))
		kbRows = append(kbRows, tgbotapi.NewInlineKeyboardRow(cb))
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, b.String())
	m.ParseMode = "Markdown"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(kbRows...)
	_, _ = h.bot.Send(m)
}

/* ---------- /cancel [id] ---------- */

func (h *Handler) handleCancelCommand(ctx context.Context, msg *tgbotapi.Message) {
	arg := strings.TrimSpace(msg.CommandArguments())
	if arg == "" {
		h.reply(msg.Chat.ID, "Формат: `/cancel <id>` или воспользуйтесь /my и нажмите «Отменить».")
		return
	}
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil || id <= 0 {
		h.reply(msg.Chat.ID, "Некорректный id.")
		return
	}

	// Админ может отменять любые брони, обычный — свои (в текущем упрощённом сервисе — без проверки владельца).
	if err := h.uc.CancelBooking(ctx, id); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	if h.isAdmin(ctx, int64(msg.From.ID)) {
		h.reply(msg.Chat.ID, "Бронь отменена администратором.")
	} else {
		h.reply(msg.Chat.ID, "Бронь отменена.")
	}
}

/* ---------- /book ---------- */

func (h *Handler) handleBookStart(ctx context.Context, msg *tgbotapi.Message) {
	rooms, err := h.uc.ListRooms(ctx)
	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	if len(rooms) == 0 {
		h.reply(msg.Chat.ID, "Нет активных переговорок.")
		return
	}

	// Клавиатура выбора комнаты
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, r := range rooms {
		data := fmt.Sprintf("book_room:%d", r.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(r.Name, data),
		))
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, "Выберите переговорку:")
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	_, _ = h.bot.Send(m)
}

/* ---------- /create_room /deactivate_room ---------- */

func (h *Handler) handleCreateRoom(ctx context.Context, msg *tgbotapi.Message) {
	if !h.isAdmin(ctx, int64(msg.From.ID)) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}
	name := strings.TrimSpace(msg.CommandArguments())
	if name == "" {
		h.reply(msg.Chat.ID, "Формат: `/create_room <name>`")
		return
	}
	if err := h.uc.AdminCreateRoom(ctx, name); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Комната создана.")
}

func (h *Handler) handleDeactivateRoom(ctx context.Context, msg *tgbotapi.Message) {
	if !h.isAdmin(ctx, int64(msg.From.ID)) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}
	arg := strings.TrimSpace(msg.CommandArguments())
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil || id <= 0 {
		h.reply(msg.Chat.ID, "Формат: `/deactivate_room <room_id>`")
		return
	}
	if err := h.uc.AdminDeleteRoom(ctx, id); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Комната деактивирована.")
}

/* ---------- callbacks (календарь/время и отмена) ---------- */

func (h *Handler) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	data := cb.Data

	switch {
	// Выбор комнаты → календарь
	case strings.HasPrefix(data, "book_room:"):
		rid, _ := strconv.ParseInt(strings.TrimPrefix(data, "book_room:"), 10, 64)
		h.bookSess[cb.From.ID] = &bookingSession{RoomID: rid}

		m := tgbotapi.NewMessage(cb.Message.Chat.ID, "Выберите день:")
		today := time.Now().In(h.cfg.OfficeTZ)
		m.ReplyMarkup = buildCalendarMarkup(today, h.cfg.OfficeTZ)
		_, _ = h.bot.Send(m)

	// Навигация по календарю
	case strings.HasPrefix(data, "calnav:"):
		d, _ := time.ParseInLocation("2006-01-02", strings.TrimPrefix(data, "calnav:"), h.cfg.OfficeTZ)
		edit := tgbotapi.NewEditMessageReplyMarkup(cb.Message.Chat.ID, cb.Message.MessageID, buildCalendarMarkup(d, h.cfg.OfficeTZ))
		_, _ = h.bot.Send(edit)

	// Выбор дня → выбор часа начала
	case strings.HasPrefix(data, "calpick:"):
		sess := h.bookSess[cb.From.ID]
		if sess == nil {
			h.reply(cb.Message.Chat.ID, "Сессия брони не найдена. Начните заново: /book")
			break
		}
		day, _ := time.ParseInLocation("2006-01-02", strings.TrimPrefix(data, "calpick:"), h.cfg.OfficeTZ)
		sess.DayLocal = time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, h.cfg.OfficeTZ)

		msg := tgbotapi.NewMessage(cb.Message.Chat.ID, fmt.Sprintf("Выберите *час начала* для %s", day.Format("02.01.2006")))
		msg.ParseMode = "Markdown"
		msg.ReplyMarkup = buildHourPicker(sess.DayLocal, "start", 0, 23)
		_, _ = h.bot.Send(msg)

	// Выбор часа (start|end)
	case strings.HasPrefix(data, "pickh:"):
		// pickh:YYYY-MM-DD:start|end:HH
		parts := strings.Split(data, ":")
		dayStr, stage, hhStr := parts[1], parts[2], parts[3]
		sess := h.bookSess[cb.From.ID]
		if sess == nil {
			h.reply(cb.Message.Chat.ID, "Сессия истекла. /book")
			break
		}
		day, _ := time.ParseInLocation("2006-01-02", dayStr, h.cfg.OfficeTZ)
		hh, _ := strconv.Atoi(hhStr)

		if stage == "start" {
			sess.StartH = &hh

			edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID,
				fmt.Sprintf("Час начала: %02d\nТеперь выберите *минуты начала*.", hh))
			edit.ParseMode = "Markdown"
			_, _ = h.bot.Send(edit)

			m := tgbotapi.NewMessage(cb.Message.Chat.ID, "Минуты:")
			m.ReplyMarkup = buildMinutePicker(day, "start", hh, 0)
			_, _ = h.bot.Send(m)
		} else {
			sess.EndH = &hh

			edit := tgbotapi.NewEditMessageText(cb.Message.Chat.ID, cb.Message.MessageID,
				fmt.Sprintf("Час окончания: %02d\nВыберите *минуты окончания*.", hh))
			edit.ParseMode = "Markdown"
			_, _ = h.bot.Send(edit)

			minStart := 0
			if sess.StartH != nil && sess.EndH != nil && *sess.EndH == *sess.StartH && sess.StartM != nil {
				minStart = *sess.StartM + 1 // строго > начала
				if minStart > 59 {
					minStart = 59
				}
			}
			m := tgbotapi.NewMessage(cb.Message.Chat.ID, "Минуты:")
			m.ReplyMarkup = buildMinutePicker(day, "end", hh, minStart)
			_, _ = h.bot.Send(m)
		}

	// Выбор минут (start|end)
	case strings.HasPrefix(data, "pickm:"):
		// pickm:YYYY-MM-DD:start|end:HH:MM
		parts := strings.Split(data, ":")
		dayStr, stage, hhStr, mmStr := parts[1], parts[2], parts[3], parts[4]
		sess := h.bookSess[cb.From.ID]
		if sess == nil {
			h.reply(cb.Message.Chat.ID, "Сессия истекла. /book")
			break
		}
		day, _ := time.ParseInLocation("2006-01-02", dayStr, h.cfg.OfficeTZ)
		hh, _ := strconv.Atoi(hhStr)
		mm, _ := strconv.Atoi(mmStr)

		if stage == "start" {
			sess.StartH, sess.StartM = &hh, &mm

			m := tgbotapi.NewMessage(cb.Message.Chat.ID, "Выберите *час окончания*")
			m.ParseMode = "Markdown"
			minHour := 0
			if sess.StartH != nil {
				minHour = *sess.StartH
			}
			m.ReplyMarkup = buildHourPicker(day, "end", minHour, 23)
			_, _ = h.bot.Send(m)
			break
		}

		// stage == "end": завершаем и создаём бронь
		sess.EndH, sess.EndM = &hh, &mm

		startLocal := time.Date(day.Year(), day.Month(), day.Day(), *sess.StartH, *sess.StartM, 0, 0, h.cfg.OfficeTZ)
		endLocal := time.Date(day.Year(), day.Month(), day.Day(), *sess.EndH, *sess.EndM, 0, 0, h.cfg.OfficeTZ)

		if !endLocal.After(startLocal) {
			h.reply(cb.Message.Chat.ID, "Время окончания должно быть *позже* начала. Попробуй снова.")
			m := tgbotapi.NewMessage(cb.Message.Chat.ID, "Выберите *час окончания*")
			m.ParseMode = "Markdown"
			m.ReplyMarkup = buildHourPicker(day, "end", *sess.StartH, 23)
			_, _ = h.bot.Send(m)
			break
		}

		// Не допускаем переход через полночь
		if endLocal.Day() != startLocal.Day() ||
			endLocal.Month() != startLocal.Month() ||
			endLocal.Year() != startLocal.Year() {
			h.reply(cb.Message.Chat.ID, "Бронь не должна переходить на следующий день. Выбери другой интервал.")
			m := tgbotapi.NewMessage(cb.Message.Chat.ID, "Выберите *час окончания* (в тот же день)")
			m.ParseMode = "Markdown"
			m.ReplyMarkup = buildHourPicker(day, "end", *sess.StartH, 23)
			_, _ = h.bot.Send(m)
			break
		}

		// Создать бронь
		cmd := usecase.CreateBookingCmd{
			RoomID: domain.RoomID(sess.RoomID),
			UserID: domain.UserID(cb.From.ID),
			Start:  startLocal.UTC(),
			End:    endLocal.UTC(),
			Note:   "",
		}
		if err := h.uc.CreateBooking(ctx, cmd); err != nil {
			h.reply(cb.Message.Chat.ID, "Не удалось создать бронь: "+err.Error())
			break
		}
		delete(h.bookSess, cb.From.ID)
		h.reply(cb.Message.Chat.ID, fmt.Sprintf("Забронировано: %s %02d:%02d–%02d:%02d",
			day.Format("02.01.2006"), *sess.StartH, *sess.StartM, *sess.EndH, *sess.EndM))

	// Кнопка «Отменить» из /my
	case strings.HasPrefix(data, "c:"):
		idStr := strings.TrimPrefix(data, "c:")
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if err := h.uc.CancelBooking(ctx, id); err != nil {
			h.reply(cb.Message.Chat.ID, "Ошибка: "+err.Error())
		} else {
			h.reply(cb.Message.Chat.ID, "Бронь отменена.")
		}

	default:
		h.reply(cb.Message.Chat.ID, "Неизвестное действие.")
	}

	// убираем «часики» на кнопке
	_, _ = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}
