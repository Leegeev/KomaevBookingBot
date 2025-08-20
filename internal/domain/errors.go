package domain

import "errors"

var (
	ErrUnauthorized       = errors.New("not authorized")
	ErrUserNotWhitelisted = errors.New("user not in whitelist")

	// room errors
	ErrRoomNotFound      = errors.New("room not found")
	ErrRoomAlreadyExists = errors.New("room already exists")
	ErrNoRoomsAvailable  = errors.New("no rooms available")

	ErrBookingNotFound       = errors.New("booking not found")
	ErrInvalidTimeRange      = errors.New("invalid time range")
	ErrPastTimeNotAllowed    = errors.New("cannot book in the past")
	ErrDurationTooShort      = errors.New("booking duration is too short")
	ErrDurationTooLong       = errors.New("booking duration is too long")
	ErrOutsideWorkingHours   = errors.New("booking outside working hours")
	ErrTimeStepViolation     = errors.New("booking does not match required time step")
	ErrOverlapsExisting      = errors.New("booking overlaps existing booking")
	ErrForbiddenCancellation = errors.New("user cannot cancel this booking")
	ErrDBConnectionFailed    = errors.New("failed to connect to database")
	ErrInvalidInputData      = errors.New("invalid input data")
	ErrNotOwner              = errors.New("user does not own this booking")
)
