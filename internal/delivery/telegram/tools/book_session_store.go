package tools

import (
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

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
	data map[int64]*bookingSession
}

func newSessionStore() *sessionsStore {
	return &sessionsStore{
		data: make(map[int64]*bookingSession),
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
