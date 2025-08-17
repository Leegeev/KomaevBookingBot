package telegram

import (
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Выбор часа (0–23)
func buildHourPicker(day time.Time, stage string, minHour, maxHour int) tgbotapi.InlineKeyboardMarkup {
	// stage: "start" | "end"
	rows := [][]tgbotapi.InlineKeyboardButton{}
	row := []tgbotapi.InlineKeyboardButton{}

	if minHour < 0 {
		minHour = 0
	}
	if maxHour > 23 {
		maxHour = 23
	}
	for h := minHour; h <= maxHour; h++ {
		data := fmt.Sprintf("pickh:%s:%s:%02d", day.Format("2006-01-02"), stage, h)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%02d", h), data))
		if len(row) == 6 {
			rows = append(rows, row)
			row = []tgbotapi.InlineKeyboardButton{}
		}
	}
	if len(row) > 0 {
		rows = append(rows, row)
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// Выбор минут (00/15/30/45), можно ограничить минимальной минутой
func buildMinutePicker(day time.Time, stage string, hour, minStart int) tgbotapi.InlineKeyboardMarkup {
	options := []int{0, 15, 30, 45}
	row := []tgbotapi.InlineKeyboardButton{}
	for _, m := range options {
		if m < minStart {
			continue
		}
		data := fmt.Sprintf("pickm:%s:%s:%02d:%02d", day.Format("2006-01-02"), stage, hour, m)
		row = append(row, tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%02d", m), data))
	}
	return tgbotapi.NewInlineKeyboardMarkup([][]tgbotapi.InlineKeyboardButton{row}...)
}
