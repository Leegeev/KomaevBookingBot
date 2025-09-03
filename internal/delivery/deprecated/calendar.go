package telegram

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func monthStart(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, d.Location())
}

func buildCalendarMarkup(current time.Time, officeTZ *time.Location) tgbotapi.InlineKeyboardMarkup {
	cur := monthStart(current.In(officeTZ))

	// Навигация
	prev := cur.AddDate(0, -1, 0)
	next := cur.AddDate(0, +1, 0)
	header := tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("«", fmt.Sprintf("calnav:%s", prev.Format("2006-01-02"))),
		tgbotapi.NewInlineKeyboardButtonData(cur.Format("January 2006"), "noop"),
		tgbotapi.NewInlineKeyboardButtonData("»", fmt.Sprintf("calnav:%s", next.Format("2006-01-02"))),
	)

	// Дни недели
	wd := tgbotapi.NewInlineKeyboardRow(
		btnNoop("Пн"), btnNoop("Вт"), btnNoop("Ср"),
		btnNoop("Чт"), btnNoop("Пт"), btnNoop("Сб"), btnNoop("Вс"),
	)

	rows := [][]tgbotapi.InlineKeyboardButton{header, wd}

	firstWD := int(cur.Weekday())
	if firstWD == 0 { // Sunday
		firstWD = 7
	}

	row := []tgbotapi.InlineKeyboardButton{}
	for i := 1; i < firstWD; i++ {
		row = append(row, btnNoop(" "))
	}

	daysInMonth := cur.AddDate(0, 1, -1).Day()
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(cur.Year(), cur.Month(), day, 0, 0, 0, 0, officeTZ)
		data := fmt.Sprintf("calpick:%s", date.Format("2006-01-02"))
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%2d", day), data))
		if len(row) == 7 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		for len(row) < 7 {
			row = append(row, btnNoop(" "))
		}
		rows = append(rows, row)
	}

	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func btnNoop(text string) tgbotapi.InlineKeyboardButton {
	return tgbotapi.NewInlineKeyboardButtonData(text, "noop")
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





	// Строка 2 — дни недели
	row2 := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	daysOfWeek := []string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
	today := time.Now()
	todayW := int(today.Weekday())
	for i := 0; i < 7; i++ {
		dayIndex := (todayW + i) % 7
		row2 = append(row2, tgbotapi.NewInlineKeyboardButtonData(daysOfWeek[dayIndex], "noop"))
	}

	// Строка 3 — конкретные даты
	row3 := make([]tgbotapi.InlineKeyboardButton, 0, 7)
	for i := 0; i < 7; i++ {
		day := today.AddDate(0, 0, i)
		display := day.Format("02.01")
		callback := fmt.Sprintf("book:calendar:%s", day.Format("2006-01-02"))
		row3 = append(row3, tgbotapi.NewInlineKeyboardButtonData(display, callback))
	}
