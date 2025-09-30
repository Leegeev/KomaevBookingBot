package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	// "github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

func (h *Handler) handleRegisterFromAdmin(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in handleRegisterFromAdmin handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
	h.log.Info("Registering group", "chatID", msg.Chat.ID, "username", msg.From.UserName)
	if msg.From.ID == h.cfg.AdminID {
		h.cfg.GroupChatID = msg.Chat.ID
		h.DailySchedule()
	}
}
