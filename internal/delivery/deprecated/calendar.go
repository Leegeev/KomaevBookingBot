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
