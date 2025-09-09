package tools

import (
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

type BookingSession struct {
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

type SessionsStore struct {
	data map[int64]*BookingSession
}

func NewSessionStore() *SessionsStore {
	return &SessionsStore{
		data: make(map[int64]*BookingSession),
	}
}

func (s *SessionsStore) Get(userID int64) *BookingSession {
	if s, ok := s.data[userID]; ok {
		return s
	}
	// newSession := &BookingSession{
	// 	UserID: userID,
	// }
	// s.data[userID] = newSession
	// return newSession
	return nil
}

func (s *SessionsStore) Set(session *BookingSession) {
	s.data[session.UserID] = session
}

func (s *SessionsStore) Delete(userID int64) {
	delete(s.data, userID)
}
