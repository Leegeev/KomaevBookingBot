package tools

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

// Step -1.
// Возвращает пустую клавиатуру для выхода в главное меню
func BuildBlankInlineKB() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.InlineKeyboardMarkup{InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{}}
}

func BuildBackInlineKBButton(data string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(TextBackInlineKBButton, data)
}

// Step 0.
// /book Строит инлайн клавиатуру с переговорками
func BuildRoomListKB(rooms []domain.Room) [][]tgbotapi.InlineKeyboardButton {
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

	// Кнопка "Назад"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(BuildBackInlineKBButton("book:list_back")))

	return rows
}

// Step 1.
// Строит календарь. Вызывается из хендлера.
func BuildCalendarKB(shift int64) tgbotapi.InlineKeyboardMarkup {
	// Навигация
	row1 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⏪", fmt.Sprintf("book:calendar_nav:%d", shift-1)),
		tgbotapi.NewInlineKeyboardButtonData("⏩", fmt.Sprintf("book:calendar_nav:%d", shift+1)),
	)

	// Определяем понедельник текущей недели
	now := time.Now()
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // воскресенье = 7
	}
	// смещаемся к понедельнику
	startOfWeek := now.AddDate(0, 0, -(weekday - 1))
	// смещаем shift недель
	startOfWeek = startOfWeek.AddDate(0, 0, int(shift*7))

	daysOfWeek := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	row2 := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	row3 := make([]tgbotapi.InlineKeyboardButton, 0, 7)

	for i := 0; i < 7; i++ {
		day := startOfWeek.AddDate(0, 0, i)

		// День недели (некликабельный)
		row2 = append(row2, tgbotapi.NewInlineKeyboardButtonData(daysOfWeek[i], "noop"))

		// Дата
		var row3display, callback string
		if shift == 0 && day.Before(now.Truncate(24*time.Hour)) {
			// прошедшие дни этой недели блокируем
			row3display = "❌"
			callback = "noop"
		} else {
			row3display = day.Format("02.01")
			callback = fmt.Sprintf("book:calendar:%s", day.Format("2006-01-02"))
		}
		row3 = append(row3, tgbotapi.NewInlineKeyboardButtonData(row3display, callback))
	}

	// Назад
	row4 := tgbotapi.NewInlineKeyboardRow(BuildBackInlineKBButton("book:calendar_back"))

	return tgbotapi.NewInlineKeyboardMarkup(row1, row2, row3, row4)
}

// Step 2.
// Строит клавиатуру для выбора длительности брони.
func BuildDurationKB() tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 4)

	for i := 1; i <= 8; i += 2 {
		// (0.5, 1.5, 2.5, 3.5)
		left := float64(i) * 0.5
		// (1.0, 2.0, 3.0, 4.0)
		right := float64(i+1) * 0.5

		leftBtn := tgbotapi.NewInlineKeyboardButtonData(
			formatDurationButtonText(left),
			fmt.Sprintf("book:duration:%.1f", left),
		)
		rightBtn := tgbotapi.NewInlineKeyboardButtonData(
			formatDurationButtonText(right),
			fmt.Sprintf("book:duration:%.1f", right),
		)

		row := tgbotapi.NewInlineKeyboardRow(leftBtn, rightBtn)
		rows = append(rows, row)
	}

	// Добавим кнопку "Назад"
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(BuildBackInlineKBButton("book:duration_back")))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildConfirmationKB() tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, 2)
	// Кнопка с
	yesBtn := tgbotapi.NewInlineKeyboardButtonData(
		"✅Верно",
		fmt.Sprintf("book:confirm:%d", 1),
	)
	noBtn := tgbotapi.NewInlineKeyboardButtonData(
		"❌Отмена",
		fmt.Sprintf("book:confirm:%d", 0),
	)
	row := tgbotapi.NewInlineKeyboardRow(yesBtn, noBtn)

	rows = append(rows, row)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(BuildBackInlineKBButton("book:confirm_back")))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func formatDurationButtonText(d float64) string {
	if d == float64(int64(d)) {
		return fmt.Sprintf("%.0f", d)
	}
	return fmt.Sprintf("%.1f", d)
}

func BuildMyListKB(bks []domain.Booking, OfficeTZ *time.Location) tgbotapi.InlineKeyboardMarkup {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(bks))
	for _, bk := range bks {
		start := bk.Range.Start.In(OfficeTZ)
		end := bk.Range.End.In(OfficeTZ)

		btnText := fmt.Sprintf("#%s — %s %02d:%02d–%02d:%02d",
			bk.RoomName, start.Format("01-02"),
			start.Hour(), start.Minute(), end.Hour(), end.Minute())

		data := fmt.Sprintf("my:list:%d", bk.ID)

		btn := tgbotapi.NewInlineKeyboardButtonData(btnText, data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(btn))
	}

	rows = append(rows, tgbotapi.NewInlineKeyboardRow(BuildBackInlineKBButton("my:back")))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildMyOperationsKB(bookingID int64) tgbotapi.InlineKeyboardMarkup {
	// rescheduleBtn := tgbotapi.NewInlineKeyboardButtonData("🔄 Перенести", fmt.Sprintf("my:reschedule:%d", bookingID))
	cancelBtn := tgbotapi.NewInlineKeyboardButtonData("❌ Отменить", fmt.Sprintf("my:cancel:%d", bookingID))

	backBtn := BuildBackInlineKBButton("my:list_back")

	// row1 := tgbotapi.NewInlineKeyboardRow(rescheduleBtn, cancelBtn)
	row1 := tgbotapi.NewInlineKeyboardRow(cancelBtn)
	row2 := tgbotapi.NewInlineKeyboardRow(backBtn)

	return tgbotapi.NewInlineKeyboardMarkup(row1, row2)
}

func BuildMainMenuKB(role string) tgbotapi.ReplyKeyboardMarkup {
	// собираем строки кнопок
	row1 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(TextMainBookButton),
		tgbotapi.NewKeyboardButton(TextMainMyButton),
	)
	row2 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(TextMainScheduleButton),
	)

	rows := [][]tgbotapi.KeyboardButton{row1, row2}

	// если админ — добавляем ещё ряд кнопок
	if CheckRoleIsAdmin(role) {
		row3 := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(TextMainCreateRoomButton),
			tgbotapi.NewKeyboardButton(TextMainDeleteRoomButton),
		)
		rows = append(rows, row3)
	}

	// собираем клавиатуру
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false

	return kb
}
