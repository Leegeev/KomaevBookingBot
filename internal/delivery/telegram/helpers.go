package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type UserID = int64

const AdminID int64 = 123456789 // TODO: заменить на реальный ID администратора

type bookingSession struct {
	BookState int64
	ChatID    int64
	UserID    int64
	MessageID int64 // сообщение, которое редактируем
	RoomID    string
	RoomName  string
	Date      time.Time // без времени, локаль офиса
	StartTime time.Time // полноценный time с датой+временем
	EndTime   time.Time
	Duration  time.Duration
}

// Roles in a chat
const (
	Creator       = "creator"
	Administrator = "administrator"
	Member        = "member"
)

/*

------callbacks------
---my---
my:select:*bk.ID*
my:back


*/

func (h *Handler) getRole(ctx context.Context, userID int64) (string, error) {
	if h.cfg.GroupChatID == 0 {
		return "", fmt.Errorf("GroupChatID is not set in config")
	}

	cfg := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: h.cfg.GroupChatID,
			UserID: userID,
		},
	}

	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in GetRole", "user_id", userID, "err", err)
		return "", err
	}

	m, err := h.bot.GetChatMember(cfg)
	if err != nil {
		h.log.Error("GetChatMember failed", "error", err)
		return "", err
	}

	return m.Status, nil
}

func (h *Handler) notifyAdmin(msg string) {
	escaped := EscapeMarkdownV2(msg)
	adminMsg := tgbotapi.NewMessage(AdminID, escaped)
	adminMsg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(adminMsg); err != nil {
		h.log.Error("Failed to notify admin", "err", err)
	}
}

// EscapeMarkdownV2 безопасно экранирует текст для Telegram MarkdownV2
func EscapeMarkdownV2(text string) string {
	var b strings.Builder

	// Список символов, требующих экранирования в MarkdownV2
	escapeChars := map[rune]bool{
		'_': true, '[': true, ']': true, '(': true, ')': true,
		'~': true, '`': true, '>': true, '#': true, '+': true, '-': true,
		'=': true, '|': true, '{': true, '}': true, '.': true, '!': true,
		'\\': true,
	}

	for _, r := range text {
		if escapeChars[r] {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}

	return b.String()
}

func getStartMessageText() string {
	return `👋 *Привет! Я бот для бронирования переговорок.*

🦾 *Вот, что я могу:*

📝 • *Забронировать* — переговорку  
📋 • *Мои брони* — показать список ваших броней  
📅 • *Расписание* — показать занятость переговорок  
📖 • *Помощь* — Подробная справка о всех командах`
}

func getHelpMessageText() string {
	return `👋 *Описание всего функционала:*

📝 • *Забронировать* — выбери удобную дату и время для встречи  
📋 • *Мои брони* — покажу список ваших броней с возможностью их *отменить* или *перенести*  
📅 • *Расписание* — покажу занятость переговорок на текущую неделю  
📖 • *Справка* — покажу это сообщение`
}

func getAdminStartMessageText() string {
	return "*Создать комнату* / *Удалить комнату* — кнопки для управления комнатами"
}

func getAdminHelpMessageText() string {
	return "🛠️ • *Создать комнату* / *Удалить комнату* — доступны и видны только администраторам чата Коллегии"
}
