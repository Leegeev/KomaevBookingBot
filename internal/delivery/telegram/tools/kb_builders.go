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
	if shift < 0 {
		shift = 0
	}

	// Строка 1 — навигация
	row1 := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⏪", fmt.Sprintf("book:calendar_nav:%d", shift-1)),
		tgbotapi.NewInlineKeyboardButtonData("⏩", fmt.Sprintf("book:calendar_nav:%d", shift+1)),
	)
	// Строка 2 — дни недели; Строка 3 — конкретные даты
	row2 := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	row3 := make([]tgbotapi.InlineKeyboardButton, 0, 7)

	now := time.Now()
	daysOfWeek := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	todayW := int(now.Weekday())                   // используется для создания календаря на ТЕКУЩЕЙ неделе
	shiftedDate := now.AddDate(0, 0, 7*int(shift)) // используется для отсчета дат

	if todayW == 0 { // То воскресенье меняем на удобный формат для daysOfWeek
		todayW = 6
	} else {
		todayW -= 1 // Если сегодня не воскресенье, то -1 чтобы даты совпадали
	}

	if shift != 0 {
		shiftedDate = shiftedDate.AddDate(0, 0, -todayW)
		todayW = -1 // Если shift != 0, значит мы строим график другой недели и там все даты доступны
	}

	var row3display, callback string
	for i := 0; i < 7; i++ {
		row2display := daysOfWeek[i]
		date := shiftedDate.AddDate(0, 0, i)
		if todayW > i {
			row3display = "❌"
			callback = ""
		} else {
			row3display = date.Format("02.01")
			callback = fmt.Sprintf("book:calendar:%s", date.Format("2006-01-02"))
		}
		row2 = append(row2, tgbotapi.NewInlineKeyboardButtonData(row2display, callback))
		row3 = append(row3, tgbotapi.NewInlineKeyboardButtonData(row3display, callback))
	}

	// Строка 4 — Назад
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
		tgbotapi.NewKeyboardButton("🗓 Забронировать"),
		tgbotapi.NewKeyboardButton("📋 Мои брони"),
	)
	row2 := tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Расписание"),
	)

	rows := [][]tgbotapi.KeyboardButton{row1, row2}

	// если админ — добавляем ещё ряд кнопок
	if CheckRoleIsAdmin(role) {
		row3 := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Создать комнату"),
			tgbotapi.NewKeyboardButton("Удалить комнату"),
		)
		rows = append(rows, row3)
	}

	// собираем клавиатуру
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	kb.OneTimeKeyboard = false

	return kb
}
