package telegram

import (
	"context"
	"fmt"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

type roleEntry struct {
	role      string
	expiresAt time.Time
}

type RoleCache struct {
	mu   sync.RWMutex
	data map[UserID]roleEntry
	ttl  time.Duration
}

func NewRoleCache(ttl time.Duration) *RoleCache {
	c := &RoleCache{
		data: make(map[UserID]roleEntry),
		ttl:  ttl,
	}
	return c
}

func (c *RoleCache) Get(id UserID) (string, bool) {
	c.mu.RLock()
	e, ok := c.data[id]
	c.mu.RUnlock()
	if !ok {
		return "", false
	}

	// проверяем срок жизни
	if time.Now().After(e.expiresAt) {
		// истёк – удаляем
		c.mu.Lock()
		delete(c.data, id)
		c.mu.Unlock()
		return "", false
	}
	return e.role, true
}

func (c *RoleCache) Set(id UserID, role string) {
	c.mu.Lock()
	c.data[id] = roleEntry{role: role, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

type UserID = int64
type BookingID = int64

func (h *Handler) getRole(userID int64) (string, error) {
	if h.cfg.GroupChatID == 0 {
		h.notifyAdmin("GroupChatID is not set in config")
		return "", fmt.Errorf("внутренняя ошибка. тех поддержка уведомлена")

	}

	// --- сначала пробуем из кеша ---
	if role, ok := h.roleCache.Get(UserID(userID)); ok {
		return role, nil
	}

	// --- если в кеше нет или TTL истёк, идём в Telegram API ---
	cfg := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: h.cfg.GroupChatID,
			UserID: userID,
		},
	}

	m, err := h.bot.GetChatMember(cfg)
	if err != nil {
		h.log.Error("GetChatMember failed", "error", err)
		return "", err
	}

	// сохраняем в кеш с TTL из конфигурации
	h.roleCache.Set(UserID(userID), m.Status)

	return m.Status, nil
}

func (h *Handler) checkSupported(ctx context.Context, upd tgbotapi.Update) error {
	var userID int64
	switch {
	case upd.Message != nil:
		userID = upd.Message.From.ID
	case upd.CallbackQuery != nil:
		userID = upd.CallbackQuery.From.ID
	default:
		return fmt.Errorf("данный update не поддерживается")
	}

	role, err := h.getRole(userID)
	if err != nil {
		return err
	}

	if !tools.CheckRoleIsSupported(role) {
		return fmt.Errorf("user is not supported")
	}
	return nil
}

func (h *Handler) notifyAdmin(msg string) {
	escaped := tools.EscapeMarkdownV2(msg)
	adminMsg := tgbotapi.NewMessage(h.cfg.AdminID, escaped)
	adminMsg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(adminMsg); err != nil {
		h.log.Error("Failed to notify admin", "err", err)
	}
}

func (h *Handler) answerWarning(warning string, cq *tgbotapi.CallbackQuery) {
	edit := tgbotapi.NewEditMessageReplyMarkup(
		cq.Message.Chat.ID,
		cq.Message.MessageID,
		tools.BuildBlankInlineKB(),
	)

	if _, err := h.bot.Send(edit); err != nil {
		h.log.Warn("Failed to edit message on confirmation", "err", err)
	}

	role, err := h.getRole(cq.From.ID)
	if err != nil {
		h.log.Warn("Failed to get user role on user", "err", err, "user_id", cq.From.ID, "username", cq.From.UserName)
		role = tools.Member
	}

	msg := tgbotapi.NewMessage(cq.Message.Chat.ID, warning)
	msg.ReplyMarkup = tools.BuildMainMenuKB(role)
	msg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(msg); err != nil {
		h.log.Error("failed to send main menu", "err", err)
	}
}
