package telegram

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

/* ---------- /create_room /deactivate_room ---------- */

// TODO: Добавить состояние сессии на создание комнаты, чтобы не приходилось писать команду и аргумент в одном сообщении.
func (h *Handler) handleCreateRoom(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in handleCreateRoom handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	role, err := h.getRole(msg.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", msg.From.ID, "username", msg.From.UserName)
		role = tools.Member
	}

	if !tools.CheckRoleIsAdmin(role) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}

	name := strings.TrimSpace(msg.CommandArguments())
	if name == "" {
		h.reply(msg.Chat.ID, tools.TextRoomNameInput.String())
		return
	}
	if err := h.uc.AdminCreateRoom(ctx, name); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Комната создана.")
}

func (h *Handler) handleDeactivateRoom(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in handleDeactivateRoom handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

	role, err := h.getRole(msg.From.ID)
	if err != nil {
		h.log.Error("Failed to get user role in deactivate room command", "user_id", msg.From.ID, "err", err)
		h.reply(msg.Chat.ID, "Ошибка при получении вашей роли")
		return
	} else if !tools.CheckRoleIsAdmin(role) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}

	rooms, err := h.uc.ListRooms(ctx)
	if errors.Is(err, domain.ErrNoRoomsAvailable) {
		h.reply(msg.From.ID, tools.TextBookNoRoomsAvailable.String())
		return
	} else if err != nil {
		h.log.Error("Failed to list rooms", "user_id", msg.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /book:* `%s`", err.Error()))
		h.reply(msg.From.ID, tools.TextBookNoRoomsErr.String())
		return
	}

	rows := tools.BuildRoomListKB(rooms, "deactivate_room")

	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextBookIntroduction.String())
	m.ParseMode = "MarkdownV2"
	m.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)

	if _, err := h.bot.Send(m); err != nil {
		h.log.Error("Failed to handle /book on rooms list", "err", err)
	}
}

func (h *Handler) handleDeactivateList(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	parts := strings.Split(cq.Data, ":")
	id, _ := strconv.ParseInt(parts[2], 10, 64)

	room, err := h.uc.GetRoom(ctx, id)
	if err != nil {
		h.reply(cq.Message.Chat.ID, "Ошибка: не удалось получить переговорку.")
		return
	}

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		string(tools.BuildRoomDeleteConfirmationSrt(room.Name)),
		tools.BuildRoomDeleteKB(id),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on handleDeactivateList", "err", err)
	}
}

func (h *Handler) handleDeactivateConfirm(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	parts := strings.Split(cq.Data, ":")
	id, _ := strconv.ParseInt(parts[2], 10, 64)

	var edit tgbotapi.EditMessageTextConfig
	if err := h.uc.AdminDeleteRoom(ctx, id); err != nil {
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при деактивации комнаты ID %d:* `%s`", id, err.Error()))
		edit = tgbotapi.NewEditMessageTextAndMarkup(
			cq.Message.Chat.ID,
			cq.Message.MessageID,
			tools.TextRoomDeactivatedErr.String(),
			tools.BuildBlankInlineKB(),
		)
	} else {
		edit = tgbotapi.NewEditMessageTextAndMarkup(
			cq.Message.Chat.ID,
			cq.Message.MessageID,
			tools.TextRoomDeactivated.String(),
			tools.BuildBlankInlineKB(),
		)
	}

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on handleDeactivateList", "err", err)
	}
}

func (h *Handler) handleConfirmCancel(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextRoomConfirmCancel.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on duration", "err", err)
	}
}

func (h *Handler) handleDeactivateListBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextRedirectingToMainMenu.String(),
		tools.BuildBlankInlineKB(),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on duration", "err", err)
	}

	role, err := h.getRole(cq.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", cq.From.ID, "username", cq.From.UserName)
		role = tools.Member
	}
	replyKB := tools.BuildMainMenuKB(role)

	msg := tgbotapi.NewMessage(cq.Message.Chat.ID, "Главное меню:")
	msg.ReplyMarkup = replyKB
	msg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Error("failed to send main menu", "err", err)
	}
}

func (h *Handler) handleDeactivateConfirmBack(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	h.answerCB(cq, "")
	h.log.Info("User clicked 'Назад' on deactivate confirm", "user_id", cq.From.ID)

	rooms, err := h.uc.ListRooms(ctx)
	if errors.Is(err, domain.ErrNoRoomsAvailable) {
		h.reply(cq.Message.From.ID, tools.TextBookNoRoomsAvailable.String())
		return
	} else if err != nil {
		h.log.Error("Failed to list rooms", "user_id", cq.Message.From.ID, "error", err)
		h.notifyAdmin(fmt.Sprintf("❗ *Ошибка при /book:* `%s`", err.Error()))
		h.reply(cq.Message.From.ID, tools.TextBookNoRoomsErr.String())
		return
	}

	rows := tools.BuildRoomListKB(rooms, "deactivate_room")

	edit := tgbotapi.NewEditMessageTextAndMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.TextBookIntroduction.String(),
		tgbotapi.NewInlineKeyboardMarkup(rows...),
	)

	edit.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(edit); err != nil {
		h.log.Error("Failed to edit message on duration", "err", err)
	}
}
