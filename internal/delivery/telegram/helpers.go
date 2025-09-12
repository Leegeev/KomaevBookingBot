package telegram

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/delivery/telegram/tools"
)

/*

	CreateBooking(ctx context.Context, cmd bookingSession) error
	CancelBooking(ctx context.Context, bookingID int64) error
	CheckBookingAndUserID(ctx context.Context, bookingID, userID int64) (bool, error)
	ListUserBookings(ctx context.Context, userID int64) ([]domain.Booking, error)
	ListRoomBookings(ctx context.Context, roomID int64) ([]domain.Booking, error)
	ListRooms(ctx context.Context) ([]domain.Room, error)
	GetRoom(ctx context.Context, roomID int64) (domain.Room, error)
	AdminCreateRoom(ctx context.Context, name string) error
	AdminDeleteRoom(ctx context.Context, roomID int64) error

*/

type UserID = int64
type BookingID = int64

func (h *Handler) getRole(userID int64) (string, error) {
	if h.cfg.GroupChatID == 0 {
		h.notifyAdmin("GroupChatID не установлен")
		return "", fmt.Errorf("GroupChatID is not set in config")
	}

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

	return m.Status, nil
}

func (h *Handler) checkSupported(ctx context.Context, upd tgbotapi.Update) error {
	if upd.Message != nil {
		role, _ := h.getRole(upd.Message.From.ID)
		supported := tools.CheckRoleIsSupported(role)
		if !supported {
			return fmt.Errorf("user is not supported")
		}
		return nil
	}

	if upd.CallbackQuery != nil {
		role, _ := h.getRole(upd.CallbackQuery.From.ID)
		supported := tools.CheckRoleIsSupported(role)
		if !supported {
			return fmt.Errorf("user is not supported")
		}
		return nil
	}
	// h.log.Error("Данный update", "upd", upd)
	return fmt.Errorf("Данный update не поддерживается")
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
