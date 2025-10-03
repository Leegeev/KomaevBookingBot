package tools

import (
	"fmt"
	"strings"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

// NOTIFIER
const (
	NotifierString = `☀️ Доброе утро, коллегия! 
📅 *Расписание переговорок на %s:*

`
	TextRegistrationSuccess      SafeText = "✅ Группа успешно зарегистрирована.\nGroupChatID сохранён."
	TextRegistrationUnauthorized SafeText = "❌ Зарегистрировать группу может только владелец. Пожалуйста, не используйте эту комманду"
)

// тексты кнопок
const (
	TextBackInlineKBButton = "🔙 Назад"

	TextMainBookButton = "📝 Забронировать"
	TextMainMyButton   = "📋 Мои бронирования"

	TextMainScheduleButton = "📅 Расписание"

	TextMainCreateRoomButton = "➕ Создать комнату"
	TextMainDeleteRoomButton = "🗑️ Удалить комнату"
	TextMainHelpButton       = "ℹ️ Помощь"
)

// тексты /start /help menu
const (
	TextMainMenu              SafeText = "🏠 Вы в главном меню"
	TextRedirectingToMainMenu SafeText = "↩️ Перенаправляю в главное меню..."
	TextStartMessage          SafeText = `👋 *Привет! Я бот для бронирования переговорок.*

🦾 *Вот, что я могу:*

📝 • *Забронировать* — переговорку  
📋 • *Мои брони* — показать список ваших броней  
📅 • *Расписание* — показать расписание переговорок  
ℹ️ • *Помощь* — справка о всех командах`

	TextHelpMessage SafeText = `👋 *Описание всего функционала:*

📝 • *Забронировать* — укажите *переговорку*, удобную *дату* и *время* для встречи. В прошлое и занятое время забронировать не получится. 
📋 • *Мои брони* — покажу список ваших броней с возможностью их *отменить*
📅 • *Расписание* — покажу расписание переговорок на следующие 7 дней
ℹ️ • *Помощь* — покажу это сообщение`
)

// тексты admin /help /start
const (
	TextAdminStartMessage SafeText = "🛠️ • *Создать комнату* / *Удалить комнату* — кнопки для управления комнатами"

	TextAdminHelpMessage SafeText = "🛠️ • *Создать комнату* / *Удалить комнату* — доступны и видны только администраторам чата Коллегии"
)

// тексты /book
const (
	TextBookIntroduction SafeText = "🏢 *Выберите переговорку:*"

	TextBookNoRoomsAvailable SafeText = `😢 На данный момент *переговорок нет*
По этому вопросу обращаться к *администрации чата коллегии*`

	TextBookNoRoomsErr SafeText = `⚠️ Что-то поломалось, тех. поддержка уже уведомлена`

	TextBookCalendar SafeText = "📅 Выберите дату:"

	TextBookAskTimeInput SafeText = `🕗 Введите начало брони:
(в формате xx, xx:00 ИЛИ xx:30)`

	TextBookTimeInvalidInput SafeText = `❌ Неверный формат времени.  
Попробуйте ещё раз в формате *чч*, *чч:00* или *чч:30*  
Например: *12*, *12:00* или *12:30*`

	TextBookAskDuration SafeText = "⏳ Выберите продолжительность:"

	TextBookAskConfirmation SafeText = `Подтвердите детали брони:
🏢 Переговорка: *%s*
📅 Дата: *%s*
🕗 Начало: *%s*
⏳ Продолжительность: *%s*`

	TextBookYes            SafeText = "🎉 Бронь успешно создана!"
	TextBookNo             SafeText = "❌ Бронь отменена."
	TextBookTooLateWaring  SafeText = "⚠️ *Нельзя создать бронь в прошлом.* Пожалуйста, выберите другое время."
	TextBookOverlapWarning SafeText = "⚠️ *В это время уже есть бронь.* Пожалуйста, попробуйте снова."
	TextBookServerError    SafeText = "⚠️ *Ошибка при создании брони.* Тех. поддержка уведомлена. Попробуйте ещё раз."
)

// тексты /my
const (
	TextMyIntroduction SafeText = "📋 *Ваши бронирования:*"
	TextMyOperations   SafeText = `🏢 Переговорка: %s
📅 Дата: *%s*
🕗 Начало: *%s*
⏳ Продолжительность: *%s*`

	TextMyBookingCancelled SafeText = "✅ Ваша бронь успешно отменена."
	TextMyBookingCancelErr SafeText = "⚠️ *Не удалось отменить бронь.* Тех. поддержка уже уведомлена."
)

// тексты /schedule
const (
	TextScheduleIntroduction SafeText = "📅 *Расписание на будущую неделю:*"
	TextWeekScheduleBooking  SafeText = `⁃%s %s:%s-%s:%s  👤 %s`
	TextTodayScheduleBooking SafeText = `⁃%s:%s-%s:%s  👤 %s`
	TextScheduleError        SafeText = "Ошибка при получении расписания, тех. поддержка уже уведомлена."
	// - мм.дд 16:30-17:30 @leegeev

)

// тексты /rooms
const (
	TextRoomNameInput SafeText = `📝 Введите название комнаты:
(Сразу после ввода названия, она будет создана)`
	TextRoomCreated SafeText = "🎉 Комната успешно создана."
	// TextRoomDeleteInput SafeText = `Введите ID комнаты для деактивации:
	TextRoomDeleteConfirmation SafeText = `❓ Вы уверены, что хотите деактивировать комнату *%s*?`
	TextRoomDeactivated        SafeText = "✅ Комната успешно удалена"
	TextRoomDeactivatedErr     SafeText = `*⚠️ Не удалось удалить комнату*. 
	Тех. поддержка уже уведомлена.`
	TextRoomConfirmCancel  SafeText = "❎ Деактивация комнаты отменена."
	TextRoomNameIsTooShort SafeText = "*⚠️ Название комнаты слишком короткое.* Минимум 2 символа."
	TextRoomNameIsTooLong  SafeText = "*⚠️ Название комнаты слишком длинное.* Максимум 50 символов."
)

func BuildRoomDeleteConfirmationSrt(name string) SafeText {
	return SafeText(fmt.Sprintf(string(TextRoomDeleteConfirmation), name))
}

func BuildWeekBookingStr(bks []domain.Booking) SafeText {
	var b strings.Builder
	for i, bk := range bks {
		if i == 0 {
			b.WriteString(fmt.Sprintf("*%s*\n", bk.RoomName))
		}
		b.WriteString(fmt.Sprintf(string(TextWeekScheduleBooking)+"\n",
			bk.Range.Start.Format("02.01"),
			bk.Range.Start.Format("15"),
			bk.Range.Start.Format("04"),
			bk.Range.End.Format("15"),
			bk.Range.End.Format("04"),
			bk.UserName,
		))
	}
	return SafeText(b.String())
}

func BuildTodayBookingStr(bks []domain.Booking) SafeText {
	var b strings.Builder
	for i, bk := range bks {
		if i == 0 {
			b.WriteString(fmt.Sprintf("*%s*\n", bk.RoomName))
		}
		b.WriteString(fmt.Sprintf(string(TextTodayScheduleBooking)+"\n",
			bk.Range.Start.Format("15"),
			bk.Range.Start.Format("04"),
			bk.Range.End.Format("15"),
			bk.Range.End.Format("04"),
			bk.UserName,
		))
	}
	return SafeText(b.String())
}

func BuildConfirmationStr(sess *BookingSession) SafeText {
	// Разложим duration
	hours := int(sess.Duration.Hours())
	minutes := int(sess.Duration.Minutes()) % 60

	var durationStr string
	if hours > 0 {
		durationStr = fmt.Sprintf("%dч", hours)
	}
	if minutes > 0 {
		if durationStr != "" {
			durationStr += " "
		}
		durationStr += fmt.Sprintf("%dмин", minutes)
	}

	return SafeText(fmt.Sprintf(
		TextBookAskConfirmation.String(),
		sess.RoomName,
		sess.Date.Format("02.01.2006"),
		sess.StartTime.Format("15:04"),
		durationStr,
	))
}

func BuildMyOperationStr(bk domain.Booking) SafeText {
	return SafeText(fmt.Sprintf(
		TextMyOperations.String(),
		bk.RoomName,
		bk.Range.Start.Format("02.01.2006"),
		bk.Range.Start.Format("15:04"),
		bk.Range.End.Sub(bk.Range.Start).String(),
	))
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
