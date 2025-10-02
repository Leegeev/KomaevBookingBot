package tools

import (
	"sync"
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

type BookingSession struct {
	BookState int64
	ChatID    int64
	UserID    int64
	UserName  string
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
	StateProccessingRoomCreation
	BookStateChoosingRoom
	BookStateChoosingDate
	BookStateChoosingStartTime
	BookStateChoosingDuration
	BookStateConfirmingBooking
)

type SessionsStore struct {
	data sync.Map // key: int64, value: *BookingSession
}

func NewSessionStore() *SessionsStore {
	return &SessionsStore{}
}

func (s *SessionsStore) Get(userID int64) *BookingSession {
	if val, ok := s.data.Load(userID); ok {
		return val.(*BookingSession)
	}
	return nil
}

func (s *SessionsStore) Set(session *BookingSession) {
	s.data.Store(session.UserID, session)
}

func (s *SessionsStore) Delete(userID int64) {
	s.data.Delete(userID)
}
