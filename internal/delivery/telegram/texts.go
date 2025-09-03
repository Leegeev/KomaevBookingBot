package telegram

import "strings"

// тексты главного меню
const (
	mainMenuText     SafeText = "Вы в главном меню"
	startMessageText SafeText = `👋 *Привет! Я бот для бронирования переговорок.*

🦾 *Вот, что я могу:*

📝 • *Забронировать* — переговорку  
📋 • *Мои брони* — показать список ваших броней  
📅 • *Расписание* — показать занятость переговорок  
📖 • *Помощь* — Подробная справка о всех командах`

	helpMessageText SafeText = `👋 *Описание всего функционала:*

📝 • *Забронировать* — выбери удобную дату и время для встречи  
📋 • *Мои брони* — покажу список ваших броней с возможностью их *отменить* или *перенести*  
📅 • *Расписание* — покажу занятость переговорок на текущую неделю  
📖 • *Справка* — покажу это сообщение`
)

// тексты для администраторов
const (
	adminStartMessageText SafeText = "🛠️ • *Создать комнату* / *Удалить комнату* — кнопки для управления комнатами"

	adminHelpMessageText SafeText = "🛠️ • *Создать комнату* / *Удалить комнату* — доступны и видны только администраторам чата Коллегии"
)

// тексты бронирования
const (
	bookIntroductionText SafeText = "*Выберите переговорку:*"

	bookNoRoomsAvailableText SafeText = `😢 На данный момент *переговорок нет*
По этому вопросу обращаться к *администрации чата коллегии*`

	bookCalendarText SafeText = "📅 Выберите дату:"

	bookAskTimeInputText SafeText = `Введите начало брони:
(в формате xx:00 ИЛИ xx:30)`

	bookAskDurationText SafeText = "🕗 Выберите продолжительность:"
)

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

type SafeText string

func (t SafeText) String() string {
	return EscapeMarkdownV2(string(t))
}
