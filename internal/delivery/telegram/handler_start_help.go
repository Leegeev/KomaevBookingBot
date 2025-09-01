package telegram

import (
	"context"
	// "fmt"
	// "strconv"
	// "strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	// "github.com/leegeev/KomaevBookingBot/internal/domain"
)

func (h *Handler) handleStart(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /start handler",
			"user", msg.From.UserName,
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
			"err", err)
		return
	}

	h.log.Info("Received /start command",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	msgText := getStartMessageText()

	role, err := h.getRole(ctx, msg.From.ID)
	if err != nil {
		h.log.Error("Failed to get user role in start command", "user_id", msg.From.ID, "err", err)
	} else {
		if role == Administrator || role == Creator {
			msgText += "\n\n" + getAdminStartMessageText()
		}
	}

	escaped := EscapeMarkdownV2(msgText)

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, escaped)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /start message",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", err)
	}
}

func (h *Handler) handleHelp(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /help handler",
			"user", msg.From.UserName,
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
			"err", err)
		return
	}

	h.log.Info("Received /help command",
		"user", msg.From.UserName,
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID)

	msgText := getHelpMessageText()

	role, err := h.getRole(ctx, msg.From.ID)
	if err != nil {
		h.log.Error("Failed to get user role in help command", "user_id", msg.From.ID, "err", err)
	} else {
		if role == Administrator || role == Creator {
			msgText += "\n\n" + getAdminHelpMessageText()
		}
	}

	escaped := EscapeMarkdownV2(msgText)

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, escaped)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /help message",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", err)
	}
}
