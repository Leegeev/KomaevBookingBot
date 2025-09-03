package telegram

import (
	"context"
	"fmt"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/leegeev/KomaevBookingBot/internal/domain"
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

const AdminID int64 = 123456789 // TODO: заменить на реальный ID администратора

// Roles in a chat
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

func (h *Handler) parseTimePick(ctx context.Context, msg string) time.Time {
	// TODO:
	return time.Now()
}

func (h *Handler) notifyAdmin(msg string) {
	escaped := EscapeMarkdownV2(msg)
	adminMsg := tgbotapi.NewMessage(AdminID, escaped)
	adminMsg.ParseMode = "MarkdownV2"

	if _, err := h.bot.Send(adminMsg); err != nil {
		h.log.Error("Failed to notify admin", "err", err)
	}
}

type bookingSession struct {
	BookState int64
	ChatID    int64
	UserID    int64
	MessageID int // сообщение, которое редактируем
	RoomID    domain.RoomID
	RoomName  string
	Date      time.Time // без времени, локаль офиса
	StartTime time.Time // полноценный time с датой+временем
	EndTime   time.Time
	Duration  time.Duration
}

const (
	StateIdle = iota
	StateProcessingCommand
	BookStateChoosingRoom
	BookStateChoosingDate
	BookStateChoosingStartTime
	BookStateChoosingDuration
	BookStateConfirmingBooking
)

type sessionsStore struct {
	data map[UserID]*bookingSession
}

func newSessionStore() *sessionsStore {
	return &sessionsStore{
		data: make(map[UserID]*bookingSession),
	}
}

func (s *sessionsStore) Get(userID int64) *bookingSession {
	if s, ok := s.data[userID]; ok {
		return s
	}
	newSession := &bookingSession{
		UserID: userID,
	}
	s.data[userID] = newSession
	return newSession
}

func (s *sessionsStore) Set(session *bookingSession) {
	s.data[session.UserID] = session
}

func (s *sessionsStore) Delete(userID int64) {
	delete(s.data, userID)
}
