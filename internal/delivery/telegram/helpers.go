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

const (
	Creator       = "creator"
	Administrator = "administrator"
	Member        = "member"
)

func (h *Handler) getRole(ctx context.Context, userID int64) (string, error) {
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

	if err := ctx.Err(); err != nil {
		h.log.Warn("Context canceled in GetRole", "user_id", userID, "err", err)
		return "", err
	}

	m, err := h.bot.GetChatMember(cfg)
	if err != nil {
		h.log.Error("GetChatMember failed", "error", err)
		return "", err
	}

	return m.Status, nil
}

func (h *Handler) checkRoleIsAdmin(role string) bool {
	return role == Administrator || role == Creator
}

func (h *Handler) checkRoleIsSupported(role string) bool {
	return role == Creator || role == Administrator || role == Member
}

func (h *Handler) checkSupported(ctx context.Context, upd tgbotapi.Update) error {
	if upd.Message != nil {
		role, _ := h.getRole(ctx, upd.Message.From.ID)
		supported := h.checkRoleIsSupported(role)
		if !supported {
			return fmt.Errorf("user is not supported")
		}
		return nil
	}

	if upd.CallbackQuery != nil {
		role, _ := h.getRole(ctx, upd.CallbackQuery.From.ID)
		supported := h.checkRoleIsSupported(role)
		if !supported {
			return fmt.Errorf("user is not supported")
		}
		return nil
	}
	h.log.Error("необработанный update", "upd", upd)
	return fmt.Errorf("неизвестный тип update, обратитесь в поддержку")
}

const AdminID int64 = 123456789 // TODO: заменить на реальный ID администратора

func (h *Handler) notifyAdmin(msg string) {
	escaped := tools.EscapeMarkdownV2(msg)
	adminMsg := tgbotapi.NewMessage(AdminID, escaped)
	adminMsg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(adminMsg); err != nil {
		h.log.Error("Failed to notify admin", "err", err)
	}
}
