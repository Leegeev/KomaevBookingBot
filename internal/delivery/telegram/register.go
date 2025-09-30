package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

func (h *Handler) handleRegisterFromAdmin(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in handleRegisterFromAdmin handler",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
	var m tgbotapi.MessageConfig
	if msg.From.ID != h.cfg.AdminID {
		m = tgbotapi.NewMessage(msg.Chat.ID, tools.TextRegistrationUnauthorized.String())
	} else {
		return
	}

	m.ParseMode = "MarkdownV2"
	if _, err := h.bot.Send(m); err != nil {
		h.log.Error("Failed to send registration confirmation", "err", err)
	}
}
