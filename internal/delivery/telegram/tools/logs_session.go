package tools

import (
	"sync"
	"time"
)

type LogsSession struct {
	State        int64
	UserID       int64
	MessageID    int
	Type         string // "sogl" или "zapros"
	Date         time.Time
	UserName     string
	Doveritel    string
	Comment      string
	Registration bool
}

// states
const (
	StateProcessingLogCreating = iota
	StateInputingName
	StateInputingDoveritel
	StageInputingComment
	StateCreateConfirming
)

type LogsStore struct {
	data sync.Map // key: int64, value: *LogsSession
}

func NewLogsStore() *LogsStore {
	return &LogsStore{}
}

func (s *LogsStore) Get(userID int64) *LogsSession {
	if val, ok := s.data.Load(userID); ok {
		return val.(*LogsSession)
	}
	return nil
}

func (s *LogsStore) Set(session *LogsSession) {
	s.data.Store(session.UserID, session)
}

func (s *LogsStore) Delete(userID int64) {
	s.data.Delete(userID)
}
