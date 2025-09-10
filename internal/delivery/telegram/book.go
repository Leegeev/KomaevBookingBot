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
	"github.com/leegeev/KomaevBookingBot/internal/usecase"
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

	rows := tools.BuildRoomListKB(rooms)

	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextBookIntroduction.String())
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
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
	h.log.Info("handling picked room in room list", "data", cq.Data, "user", cq.From.UserName)

	parts := strings.Split(cq.Data, ":")
	id, _ := strconv.ParseInt(parts[2], 10, 64)

	room, err := h.uc.GetRoom(ctx, id)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: не удалось получить переговорку.")
		return
	}

	// Создаем bookingSession и сохраняем в in-memory storage
	h.sessions.Set(&tools.BookingSession{
		BookState: tools.BookStateChoosingDate,
		ChatID:    cq.Message.Chat.ID,
		UserID:    cq.From.ID,
		MessageID: cq.Message.MessageID,
		RoomID:    room.ID,
		RoomName:  room.Name,
		Date:      time.Now().In(h.cfg.OfficeTZ).Truncate(24 * time.Hour),
	})

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextBookCalendar.String(),
		tools.BuildCalendarKB(0),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on book list", "err", err)
	}
}

// Step 0.+-1
// Хендлер Навигации по календарю
func (h *Handler) handleBookCalendarNavigation(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.log.Info("calendar nav callback", "data", cq.Data, "user", cq.From.UserName)

	// 1. Парсим направление навигации
	parts := strings.Split(cq.Data, ":")
	shift, _ := strconv.ParseInt(parts[2], 10, 64) // -3 -2 -1 1 2 3

	// 2. Обновляем только клавиатуру
	editMarkup := tgbotapi.NewEditMessageReplyMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildCalendarKB(shift),
	)

	if _, err := h.bot.Request(editMarkup); err != nil {
		h.log.Error("failed to edit calendar inline keyboard", "err", err)
	}
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
	session.BookState = tools.BookStateChoosingStartTime

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextBookAskTimeInput.String(),
		tgbotapi.NewInlineKeyboardMarkup(
			[]tgbotapi.InlineKeyboardButton{
				tools.BuildBackInlineKBButton("book:timepick_back"),
			}),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on calendar", "err", err)
	}
}

// Step 2.
// Обработчик РУЧНОГО ввода времени.
func (h *Handler) handleBookTimepick(ctx context.Context, msg *tgbotapi.Message) {
	h.log.Info("Received users book time input",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	startTime, err := tools.ParseTimePick(msg.Text)
	if err != nil {
		reply := tgbotapi.NewMessage(msg.Chat.ID, tools.TextBookTimeInvalidInput.String())
		reply.ParseMode = "MarkdownV2"

		if _, sendErr := h.bot.Send(reply); sendErr != nil {
			h.log.Error("Failed to send invalid time format message", "err", sendErr)
		}
		return
	}

	session := h.sessions.Get(msg.From.ID)
	if session == nil {
		h.reply(msg.From.ID, "Сессия не найдена")
		return
	}
	session.BookState = tools.BookStateChoosingDuration
	session.StartTime = startTime

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		msg.Chat.ID,
		session.MessageID,
		tools.TextBookAskTimeInput.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on timepick", "err", err)
	}

	newMsg := tgbotapi.NewMessage(msg.Chat.ID, tools.TextBookAskDuration.String())
	newMsg.ParseMode = "MarkdownV2"
	newMsg.ReplyMarkup = tools.BuildDurationKB()

	if _, err := h.bot.Send(newMsg); err != nil {
		h.log.Error("Failed to send a new message on timepick", "err", err)
	}
}

func (h *Handler) handleBookDuration(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	durStr := parts[2]
	// (0.5, 1.5, 2.5, 3.5)
	// (1.0, 2.0, 3.0, 4.0)

	durF, err := strconv.ParseFloat(durStr, 64)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: неправильная длительность")
		return
	}

	session := h.sessions.Get(cq.From.ID)
	if session == nil {
		h.reply(cq.Message.Chat.ID, "Сессия не найдена")
		return
	}

	session.Duration = time.Duration(durF * float64(time.Hour))
	session.EndTime = session.StartTime.Add(session.Duration)
	session.BookState = tools.BookStateConfirmingBooking

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildConfirmationStr(session),
		tools.BuildConfirmationKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on duration", "err", err)
	}
}

func (h *Handler) handleBookConfirm(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	defer h.sessions.Delete(cq.From.ID)

	h.answerCB(cq, "")

	parts := strings.Split(cq.Data, ":")
	confirm, _ := strconv.ParseInt(parts[2], 10, 64) // 0 || 1

	session := h.sessions.Get(cq.From.ID)

	sess := usecase.CreateBookingCmd{
		RoomID:   session.RoomID,
		RoomName: session.RoomName,
		UserID:   domain.UserID(session.UserID),
		Start:    session.StartTime,
		End:      session.EndTime,
	}

	text := ""

	if confirm == 1 {
		err := h.uc.CreateBooking(ctx, sess)
		if err != nil {
			h.reply(cq.Message.Chat.ID, "Ошибка при создании брони: "+err.Error())
			return
		}
		text = tools.TextBookYes.String()
	} else {
		text = tools.TextBookNo.String()
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildConfirmationStr(session),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on confirmation", "err", err)
	}
	newMsg := tgbotapi.NewMessage(cq.Message.Chat.ID, text)
	newMsg.ParseMode = "MarkdownV2"
	// newMsg.ReplyMarkup = tools.BuildDurationKB()

	if _, err := h.bot.Send(newMsg); err != nil {
		h.log.Error("Failed to send a new message on confirmation", "err", err)
	}
}
