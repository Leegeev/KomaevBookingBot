package telegram

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

func (h *Handler) handleLog(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLog",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
	kb := tools.BuildLogMainKB()
	m := tgbotapi.NewMessage(msg.Chat.ID, tools.TextLogMainMenu.String())
	m.ReplyMarkup = kb
	m.ParseMode = "MarkdownV2"
	go func() {
		if _, err := h.bot.Send(m); err != nil {
			h.log.Error("Failed to send a new message on handleLog", "err", err)
		}
	}()
}

func (h *Handler) handleLogFind(ctx context.Context, msg *tgbotapi.Message) {
	// TODO
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogFind",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
}

func (h *Handler) handleLogSoglasheniya(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogSoglasheniya",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

}

func (h *Handler) handleLogZaprosi(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogZaprosi",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
}

func (h *Handler) handleLogCreate(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogCreate",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
}

func (h *Handler) handleLogExport(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogExport",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}

}

func (h *Handler) handleLogMy(ctx context.Context, msg *tgbotapi.Message) {
	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in /handleLogMy",
			"user", msg.From.UserName,
			"chat_id", msg.Chat.ID,
			"err", ctx.Err())
		return
	}
}
