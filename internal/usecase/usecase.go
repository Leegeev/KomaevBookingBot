package usecase

import (
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type BookingService struct {
	roomRepo    domain.RoomRepository
	userRepo    domain.UserRepository
	bookingRepo domain.BookingRepository
	logger      logger.Logger
}

func NewBookingService(roomRepo domain.RoomRepository, userRepo domain.UserRepository, bookingRepo domain.BookingRepository, logger logger.Logger) *BookingService {
	return &BookingService{
		roomRepo:    roomRepo,
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
		logger:      logger,
	}
}

func (s *BookingService) CreateBooking(booking domain.Booking) error {
	return nil
}

func (s *BookingService) CancelBooking(bookingID int64) error {
	return nil
}

func (s *BookingService) AdminCancelBooking(bookingID int64) error {
	return nil
}

func (s *BookingService) ListUserBookings(userID int64) ([]domain.Booking, error) {
	return nil, nil
}

func (s *BookingService) ListRoomBookings(roomID int64) ([]domain.Booking, error) {
	return nil, nil
}

func (s *BookingService) ListRooms() ([]domain.Room, error) {
	return nil, nil
}

func (s *BookingService) AdminCreateRoom(room domain.Room) error {
	return nil
}

func (s *BookingService) AdminDeleteRoom(roomID int64) error {
	return nil
}

// getChatMember). Если status ∈ {creator, administrator, member} — добавляем/обновляем в users_whitelist
