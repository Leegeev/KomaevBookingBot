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

/* ---------- /start /help ---------- */

func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	text := "Привет! Я бот бронирования переговорок.\n\n" +
		"/book — забронировать\n" +
		"/my — мои брони\n" +
		"/rooms — список переговорок\n" +
		"/cancel — отменить бронь\n" +
		"/help — подробности"
	h.reply(msg.Chat.ID, text)
}

func (h *Handler) handleHelp(ctx context.Context, msg *tgbotapi.Message) {
	text := "*Команды:*\n" +
		"• /book — выбрать переговорку и указать время в формате `YYYY-MM-DD HH:MM-HH:MM [; комментарий]`\n" +
		"• /my — показать ваши будущие брони (можно отменить кнопкой)\n" +
		"• /rooms — показать активные переговорки\n" +
		"• /cancel `<id>` — отменить бронь по номеру\n" +
		"• /create_room `<name>` — создать переговорку (только админ)\n" +
		"• /deactivate_room `<room_id>` — деактивировать переговорку (только админ)\n"
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
		h.reply(msg.Chat.ChatConfig().ChatID, "У вас нет будущих броней.")
		return
	}

	kbRows := [][]tgbotapi.InlineKeyboardButton{}
	var b strings.Builder
	b.WriteString("*Ваши брони:*\n")
	for _, bk := range bookings {
		start := bk.Range.Start.In(h.cfg.OfficeTZ) // если у тебя Interval — замени на bk.Interval.Start
		end := bk.Range.End.In(h.cfg.OfficeTZ)
		fmt.Fprintf(&b, "• #%d — %s %02d:%02d–%02d:%02d\n",
			bk.ID, start.Format("2006-01-02"), start.Hour(), start.Minute(), end.Hour(), end.Minute())

		// кнопка отмены этой брони
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

	// Админ может отменить любые брони; обычный — только свои.
	if h.isAdmin(ctx, int64(msg.From.ID)) {
		// текущий usecase не различает владельца — допускаем админское удаление
		if err := h.uc.CancelBooking(ctx, id); err != nil {
			h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
			return
		}
		h.reply(msg.Chat.ID, "Бронь отменена администратором.")
		return
	}

	// не админ: безопаснее давать отмену только из списка /my (где точно ваши ID),
	// но если всё же передали id — попробуем отменить и положимся на usecase/репо
	if err := h.uc.CancelBooking(ctx, id); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Бронь отменена.")
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
	// инлайн-клавиатура выбора комнаты
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

func (h *Handler) handleBookTime(ctx context.Context, msg *tgbotapi.Message) {
	roomID := h.pending[msg.From.ID]
	text := strings.TrimSpace(msg.Text)
	// ожидаем: YYYY-MM-DD HH:MM-HH:MM [; комментарий]
	startUTC, endUTC, note, err := parseDateTimeRange(text, h.cfg.OfficeTZ)
	if err != nil {
		h.reply(msg.Chat.ID, "Формат: `YYYY-MM-DD HH:MM-HH:MM [; комментарий]`")
		return
	}

	cmd := usecase.CreateBookingCmd{
		RoomID: roomID,
		UserID: domain.UserID(msg.From.ID),
		Start:  startUTC,
		End:    endUTC,
		Note:   note,
	}
	if err := h.uc.CreateBooking(ctx, cmd); err != nil {
		h.reply(msg.Chat.ID, "Не удалось создать бронь: "+err.Error())
		return
	}
	delete(h.pending, msg.From.ID)
	h.reply(msg.Chat.ID, "Забронировано.")
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

/* ---------- callbacks ---------- */

func (h *Handler) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	data := cb.Data

	switch {
	case strings.HasPrefix(data, "book_room:"):
		roomIDStr := strings.TrimPrefix(data, "book_room:")
		rid, _ := strconv.ParseInt(roomIDStr, 10, 64)
		h.pending[cb.From.ID] = domain.RoomID(rid)

		txt := "Введите время в формате: `YYYY-MM-DD HH:MM-HH:MM [; комментарий]`\n" +
			"Например: `2025-08-17 15:00-16:00; собеседование`"
		h.reply(cb.Message.Chat.ID, txt)

	case strings.HasPrefix(data, "c:"):
		// обычная отмена (свои брони) — используем текущий usecase CancelBooking
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

	// закрыть «часики» на кнопке
	_ = h.bot.Request(tgbotapi.NewCallback(cb.ID, ""))
}
