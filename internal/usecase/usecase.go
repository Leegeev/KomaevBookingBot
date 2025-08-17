package usecase

import (
	"context"
	"time"

	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

func NewBookingService(roomRepo domain.RoomRepository, bookingRepo domain.BookingRepository, logger logger.Logger) *BookingService {
	return &BookingService{
		roomRepo:    roomRepo,
		bookingRepo: bookingRepo,
		logger:      logger,
		// userRepo:    userRepo,
	}
}

type BookingService struct {
	roomRepo    domain.RoomRepository
	bookingRepo domain.BookingRepository
	logger      logger.Logger
	// userRepo    domain.UserRepository
}

type CreateBookingCmd struct {
	RoomID domain.RoomID
	UserID domain.UserID
	Start  time.Time // UTC
	End    time.Time // UTC
	Note   string
}

func (s *BookingService) CreateBooking(ctx context.Context, cmd CreateBookingCmd) error {
	s.logger.Info("Creating booking", "booking", cmd)

	// Validate input
	// Create TimeRange
	tr, err := domain.NewTimeRange(cmd.Start, cmd.End)
	if err != nil {
		s.logger.Error("Invalid time range", "error", err)
		return err
	}

	// Create booking entity
	booking, err := domain.NewBooking(cmd.RoomID, cmd.UserID, tr, cmd.Note)
	if err != nil {
		s.logger.Error("Failed to create booking entity", "error", err)
		return err
	}

	// Save booking to repository
	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		s.logger.Error("Failed to create booking", "error", err)
		return err
	}
	return nil
}

func (s *BookingService) CancelBooking(ctx context.Context, bookingID int64) error {
	s.logger.Info("Canceling booking", "bookingID", bookingID)
	if bookingID <= 0 {
		s.logger.Error("Invalid booking ID", "bookingID", bookingID)
		return domain.ErrInvalidInputData
	}
	if err := s.bookingRepo.Delete(ctx, domain.BookingID(bookingID)); err != nil {
		s.logger.Error("Failed to cancel booking", "error", err)
		return err
	}
	return nil
}

func (s *BookingService) ListUserBookings(ctx context.Context, userID int64) ([]domain.Booking, error) {
	s.logger.Info("Listing bookings for user", "userID", userID)
	if userID <= 0 {
		s.logger.Error("Invalid user ID", "userID", userID)
		return nil, domain.ErrInvalidInputData
	}
	bookings, err := s.bookingRepo.ListByUser(ctx, domain.UserID(userID), domain.MustUTC(time.Now().UTC()))
	if err != nil {
		s.logger.Error("Failed to list user bookings", "error", err)
		return nil, err
	}
	s.logger.Info("Found bookings for user", "userID", userID, "count", len(bookings))
	return bookings, nil
}

func (s *BookingService) ListRoomBookings(ctx context.Context, roomID int64) ([]domain.Booking, error) {
	s.logger.Info("Listing bookings for room", "roomID", roomID)
	if roomID <= 0 {
		s.logger.Error("Invalid room ID", "roomID", roomID)
		return nil, domain.ErrInvalidInputData
	}
	bookings, err := s.bookingRepo.ListByRoomAndInterval(ctx, domain.RoomID(roomID), time.Now().UTC(), time.Now().UTC().Add(time.Hour*24*7))
	if err != nil {
		s.logger.Error("Failed to list room bookings", "error", err)
		return nil, err
	}
	s.logger.Info("Found bookings for room", "roomID", roomID, "count", len(bookings))
	return bookings, nil
}

func (s *BookingService) ListRooms(ctx context.Context) ([]domain.Room, error) {
	s.logger.Info("Listing all rooms")
	rooms, err := s.roomRepo.List(ctx)
	if err != nil {
		s.logger.Error("Failed to list rooms", "error", err)
		return nil, err
	}
	s.logger.Info("Found rooms", "count", len(rooms))
	if len(rooms) == 0 {
		s.logger.Warn("No rooms found")
		return nil, domain.ErrNoRoomsAvailable
	}
	return rooms, nil
}

func (s *BookingService) GetRoom(ctx context.Context, roomID int64) (domain.Room, error) {
	s.logger.Info("Getting room", "roomID", roomID)
	if roomID <= 0 {
		s.logger.Error("Invalid room ID", "roomID", roomID)
		return domain.Room{}, domain.ErrInvalidInputData
	}
	room, err := s.roomRepo.Get(ctx, domain.RoomID(roomID))
	if err != nil {
		s.logger.Error("Failed to get room", "error", err)
		return domain.Room{}, err
	}

	s.logger.Info("Room found", "roomID", roomID, "name", room.Name)
	return room, nil
}

// func (s *BookingService) AdminCancelBooking(ctx context.Context, bookingID int64) error {
// 	return nil
// }

func (s *BookingService) AdminCreateRoom(ctx context.Context, name string) error {
	if name == "" {
		s.logger.Error("Room name is empty")
		return domain.ErrInvalidInputData
	}
	s.logger.Info("Creating room", "name", name)
	room := domain.Room{Name: name}
	if _, err := s.roomRepo.Create(ctx, room); err != nil {
		s.logger.Error("Failed to create room", "error", err)
		return err
	}
	s.logger.Info("Room created successfully", "name", name)
	return nil
}

func (s *BookingService) AdminDeleteRoom(ctx context.Context, roomID int64) error {
	s.logger.Info("Deleting room", "roomID", roomID)
	if roomID <= 0 {
		s.logger.Error("Invalid room ID", "roomID", roomID)
		return domain.ErrInvalidInputData
	}
	if err := s.roomRepo.Delete(ctx, domain.RoomID(roomID)); err != nil {
		s.logger.Error("Failed to delete room", "error", err)
		return err
	}
	s.logger.Info("Room deleted successfully", "roomID", roomID)
	return nil
}

// getChatMember). Если status ∈ {creator, administrator, member} — добавляем/обновляем в users_whitelist
