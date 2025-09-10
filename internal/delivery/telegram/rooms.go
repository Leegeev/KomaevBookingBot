package telegram

import (
	"context"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

/* ---------- /create_room /deactivate_room ---------- */

func (h *Handler) handleCreateRoom(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /CreateRoom handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	}
	h.log.Info("Received /create_room command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)

	role, err := h.getRole(ctx, msg.From.ID)
	if err != nil {
		h.log.Error("Failed to get user role in create room command", "user_id", msg.From.ID, "err", err)
		h.reply(msg.Chat.ID, "Ошибка при получении вашей роли")
		return
	} else if !tools.CheckRoleIsAdmin(role) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}

	name := strings.TrimSpace(msg.CommandArguments())
	if name == "" {
		h.reply(msg.Chat.ID, "Формат: `/create_room <name>`")
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
		h.log.Warn("Context canceled in /DeleteRoom handler", "user", msg.From.UserName, "chat_id", msg.Chat.ID, "err", ctx.Err())
		return
	}
	h.log.Info("Received /deactivate_room command", "user", msg.From.UserName, "chat_id", msg.Chat.ID)

	role, err := h.getRole(ctx, msg.From.ID)
	if err != nil {
		h.log.Error("Failed to get user role in deactivate room command", "user_id", msg.From.ID, "err", err)
		h.reply(msg.Chat.ID, "Ошибка при получении вашей роли")
		return
	} else if !tools.CheckRoleIsAdmin(role) {
		h.reply(msg.Chat.ID, "Недостаточно прав.")
		return
	}

	arg := strings.TrimSpace(msg.CommandArguments())
	id, err := strconv.ParseInt(arg, 10, 64)
	if err != nil || id <= 0 {
		h.reply(msg.Chat.ID, "Формат: `/deactivate_room <room_id>`")
		return
	}
	if err := h.uc.AdminDeleteRoom(ctx, id); err != nil {
		h.reply(msg.Chat.ID, "Ошибка: "+err.Error())
		return
	}
	h.reply(msg.Chat.ID, "Комната деактивирована.")
}
