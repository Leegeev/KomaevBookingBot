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

/* ---------- /book ---------- */
// остановился тут
func (h *Handler) handleBookStart(ctx context.Context, msg *tgbotapi.Message) {
	select {
	case <-ctx.Done():
		h.log.Warn("Context canceled in /book handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	default:
	}

	h.log.Info("Received /book command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)
	rooms, err := h.uc.ListRooms(ctx)
	if err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}

	// Клавиатура выбора комнаты
	rows := [][]tgbotapi.InlineKeyboardButton{}
	for _, r := range rooms {
		data := fmt.Sprintf("book_room:%s", r.ID)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(r.Name, data),
		))
	}
	m := tgbotapi.NewMessage(msg.Chat.ID, "Выберите переговорку:")
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	_, _ = h.bot.Send(m)
}

/* ---------- callbacks (календарь/время и отмена) ---------- */

func (h *Handler) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	h.log.Info("Received callback", "user", cb.From.UserName, "chat_id", cb.Message.Chat.ID, "data", cb.Data)
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
