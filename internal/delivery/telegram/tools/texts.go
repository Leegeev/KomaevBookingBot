package tools

import (
	"fmt"
	"strings"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

const TextBackInlineKBButton = "🔙 Назад"

// тексты главного меню
const (
	TextMainMenu     SafeText = "Вы в главном меню"
	TextStartMessage SafeText = `👋 *Привет! Я бот для бронирования переговорок.*

🦾 *Вот, что я могу:*

📝 • *Забронировать* — переговорку  
📋 • *Мои брони* — показать список ваших броней  
📅 • *Расписание* — показать занятость переговорок  
📖 • *Помощь* — Подробная справка о всех командах`

	TextHelpMessage SafeText = `👋 *Описание всего функционала:*

📝 • *Забронировать* — выбери удобную дату и время для встречи  
📋 • *Мои брони* — покажу список ваших броней с возможностью их *отменить* или *перенести*  
📅 • *Расписание* — покажу занятость переговорок на текущую неделю  
📖 • *Справка* — покажу это сообщение`
)

// тексты для администраторов
const (
	TextAdminStartMessage SafeText = "🛠️ • *Создать комнату* / *Удалить комнату* — кнопки для управления комнатами"

	TextAdminHelpMessage SafeText = "🛠️ • *Создать комнату* / *Удалить комнату* — доступны и видны только администраторам чата Коллегии"
)

// тексты бронирования
const (
	TextBookIntroduction SafeText = "*Выберите переговорку:*"

	TextBookNoRoomsAvailable SafeText = `😢 На данный момент *переговорок нет*
По этому вопросу обращаться к *администрации чата коллегии*`

	TextBookNoRoomsErr SafeText = `😢 Что-то поломалось, тех. поддержка уже уведомлена`

	TextBookCalendar SafeText = "📅 Выберите дату:"

	TextBookAskTimeInput SafeText = `Введите начало брони:
(в формате xx:00 ИЛИ xx:30)`

	TextBookTimeInvalidInput SafeText = `❌ Неверный формат времени.\n
Пожалуйста, введите время в формате xx:00 ИЛИ xx:30
(Например 12:00 или 12:30)`

	TextBookAskDuration SafeText = "🕗 Выберите продолжительность:"

	TextBookAskConfirmation SafeText = `Подтвердите детали брони:
Переговорка: %s
Дата: %s
Начало: %s
Продолжительность: %s`

	TextBookYes SafeText = "✅ Бронь успешно создана!"
	TextBookNo  SafeText = "❌ Бронь отменена."
)

// ОСТАНОВИЛСЯ ЗДЕСЬ
// ДЕЛАЮ handleMyList
// тексты для мои брони
const (
	TextMyIntroduction SafeText = "*Ваши брони:*"
	TextMyOperations   SafeText = `Переговорка: %s
Дата: %s
Начало: %s
Продолжительность: %s`

	TextMyBookingCancelled SafeText = "✅ Ваша бронь успешно отменена."
	TextMyBookingCancelErr SafeText = "❌ Не удалось отменить бронь. Тех. поддержка уже уведомлена."
)

func BuildConfirmationStr(sess *BookingSession) string {
	return fmt.Sprintf(
		TextBookAskConfirmation.String(),
		sess.RoomName,
		sess.Date.Format("02.01.2006"),
		sess.StartTime.Format("15:04"),
		sess.Duration,
	)
}

func BuildMyOperationStr(bk domain.Booking) string {
	return fmt.Sprintf(
		TextMyOperations.String(),
		bk.RoomName,
		bk.Range.Start.Format("02.01.2006"),
		bk.Range.Start.Format("15:04"),
		bk.Range.End.Sub(bk.Range.Start).String(),
	)
}

type SafeText string

func (t SafeText) String() string {
	return EscapeMarkdownV2(string(t))
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
