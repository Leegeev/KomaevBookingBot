package telegram

import (
	"context"
	// "fmt"
	// "strconv"
	// "strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
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

	msgText := tools.TextStartMessage.String()

	role, err := h.getRole(msg.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", msg.From.ID, "username", msg.From.UserName)
		role = tools.Member
	}
	if tools.CheckRoleIsAdmin(role) {
		msgText += "\n\n" + tools.TextAdminStartMessage.String()
	}

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, msgText)
	msgOut.ReplyMarkup = tools.BuildMainMenuKB(role)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /start message",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", err)
	}
}

func (h *Handler) handleMainMenu(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleMainMenu handler",
			"user", msg.From.UserName,
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
			"err", err)
		return
	}

	msgText := tools.TextRedirectingToMainMenu.String()

	role, err := h.getRole(msg.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", msg.From.ID, "username", msg.From.UserName)
		role = tools.Member
	}
	if tools.CheckRoleIsAdmin(role) {
		msgText += "\n\n" + tools.TextAdminStartMessage.String()
	}

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, msgText)
	msgOut.ReplyMarkup = tools.BuildMainMenuKB(role)
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

	msgText := tools.TextHelpMessage.String()

	role, err := h.getRole(msg.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", msg.From.ID, "username", msg.From.UserName)
		role = tools.Member
	}
	if tools.CheckRoleIsAdmin(role) {
		msgText += "\n\n" + tools.TextAdminHelpMessage.String()
	}

	msgOut := tgbotapi.NewMessage(msg.Chat.ID, msgText)
	msgOut.ReplyMarkup = tools.BuildMainMenuKB(role)
	msgOut.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msgOut); err != nil {
		h.log.Error("Failed to send /help message",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", err)
	}
}
