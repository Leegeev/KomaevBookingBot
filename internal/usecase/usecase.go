package usecase

import (
	"github.com/leegeev/KomaevBookingBot/internal/domain"
	"github.com/leegeev/KomaevBookingBot/pkg/logger"
)

type OrderService struct {
	roomRepo    domain.RoomRepository
	userRepo    domain.UserRepository
	bookingRepo domain.BookingRepository
	logger      logger.Logger
}

func NewOrderService(roomRepo domain.RoomRepository, userRepo domain.UserRepository, bookingRepo domain.BookingRepository, logger logger.Logger) *OrderService {
	return &OrderService{
		roomRepo:    roomRepo,
		userRepo:    userRepo,
		bookingRepo: bookingRepo,
		logger:      logger,
	}
}
