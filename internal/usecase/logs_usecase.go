package usecase

import (
	"context"
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
)

// func (s *BookingService) CreateBooking(ctx context.Context, cmd CreateBookingCmd) error {
func (s *LogService) GetUser(ctx context.Context, cmd CreateLogCmd) error {
	// TODO
	return nil
}
func (s *LogService) CreateUser(ctx context.Context, cmd CreateLogCmd) error {
	// TODO:
	return nil
}

// GetSoglasheniyaByUserID
// GetZaprosiByUserId
// CreateLog // по type определить какую запись создать

type CreateLogCmd struct {
	UserID    domain.UserID
	UserName  string
	Type      string // "sogl" или "zapros"
	Date      time.Time
	Doveritel string
	Comment   string
}
