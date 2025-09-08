package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

// Step 0.
// /book
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
		h.reply(msg.From.ID, tools.TextBookNoRoomsAvailable.String())
		return
	} else if err != nil {
		h.log.Error("Failed to list rooms", "user_id", msg.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /book:* `%s`", err.Error()))
		h.reply(msg.From.ID, tools.TextBookNoRoomsErr.String())
		return
	}

	rows := tools.BuildRoomListKB(ctx, rooms)

	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextBookIntroduction.String())
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if _, err := h.bot.Send(m); err != nil {
		h.log.Error("Failed to handle /book on rooms list", "err", err)
	}
}

// Step 0.
// Хендлер выбора переговорки
// Строит календарь
func (h *Handler) handleBookList(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	id, _ := strconv.ParseInt(parts[2], 10, 64)

	room, err := h.uc.GetRoom(ctx, id)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: не удалось получить переговорку.")
		return
	}

	// Создаем bookingSession и сохраняем в in-memory storage
	h.sessions.Set(&tools.BookingSession{
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
		tools.TextBookCalendar.String(),
		tools.BuildCalendarKB(time.Now()),
	)

	msg.ParseMode = "MarkdownV2"
	_, _ = h.bot.Send(msg)
}

// Step 1.
// Парсит callback Дату из календаря (получает дату в local tz)
// Редактирует сообщение на ввод времени
func (h *Handler) handleBookCalendar(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	dateStr := parts[2] // формат даты предполагается как "YYYY-MM-DD"

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

	session.Date = date // LOCAL TIMEZONE

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextBookAskTimeInput.String(),
		tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				tgbotapi.NewInlineKeyboardButtonData("⬅ Назад", "book:timepick_back"),
			}),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on calendar back", "err", err)
	}
}

// Step 2.
// Обработчик РУЧНОГО ввода времени.
func (h *Handler) handleBookTimepick(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /my handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	h.log.Info("Received users book time input",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)
	session := h.sessions.Get(msg.From.ID)

	if session == nil {
		h.reply(msg.Chat.ID, "Сессия не найдена")
		return
	}

	session.StartTime = h.parseTimePick(ctx, msg.Text)

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		msg.Chat.ID,
		msg.MessageID,
		tools.TextBookAskDuration.String(),
		tools.BuildDurationKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on calendar back", "err", err)
	}
}

func (h *Handler) handleBookDuration(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	durStr := parts[2] // формат даты предполагается как "YYYY-MM-DD"

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

func (h *Handler) handleBookConfirm(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	confirm, _ := strconv.ParseInt(parts[2], 10, 64) // 0 || 1

	session := h.sessions.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	if confirm == 1 {
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
