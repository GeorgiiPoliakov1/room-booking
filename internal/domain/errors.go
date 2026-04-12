package domain

import (
	"errors"
	"fmt"
)

var (
	ErrRoomNameRequired   = errors.New("room name is required")
	ErrRoomNameTooLong    = fmt.Errorf("room name too long (max %d characters)", MaxRoomNameLength)
	ErrDescriptionTooLong = fmt.Errorf("description too long (max %d characters)", MaxDescriptionLength)
	ErrCapacityOutOfRange = fmt.Errorf("capacity must be between %d and %d", MinCapacity, MaxCapacity)
	ErrRoomNotFound       = errors.New("room not found")
	ErrRoomNameExists     = errors.New("room name already exists")
	ErrInvalidRoomID      = errors.New("room ID must be a valid non-empty UUID")

	ErrScheduleAlreadyExists = errors.New("schedule for this room already exists and cannot be changed")
	ErrEmptyDaysOfWeek       = errors.New("at least one day of week must be specified")
	ErrInvalidDayOfWeek      = errors.New("day of week must be between 1 and 7")
	ErrInvalidTimeFormat     = errors.New("time must be in HH:MM format (24h)")
	ErrEndTimeBeforeStart    = errors.New("end time must be after start time")
	ErrSlotTooShort          = errors.New("slot duration must be at least 30 minutes")
	ErrInvalidScheduleID     = errors.New("schedule ID must be a valid non-empty UUID")
	ErrDuplicateDayOfWeek    = errors.New("daysOfWeek must not contain duplicate values")

	ErrInvalidSlotDuration = errors.New("slot duration must be exactly 30 minutes")
	ErrSlotNotUTC          = errors.New("slot times must be in UTC")

	ErrBookingNotFound         = errors.New("booking not found")
	ErrSlotAlreadyBooked       = errors.New("slot is already booked")
	ErrBookingAlreadyCancelled = errors.New("booking is already cancelled")
	ErrInvalidBookingID        = errors.New("invalid booking ID")
	ErrInvalidSlotID           = errors.New("invalid slot ID")
	ErrInvalidUserID           = errors.New("invalid user ID")
	ErrInvalidBookingStatus    = errors.New("invalid booking status, must be 'active' or 'cancelled'")
	ErrCannotChangePastBooking = errors.New("cannot modify or cancel booking after slot end time")
	ErrSlotNotFound            = errors.New("slot not found")
	ErrSlotInPast              = errors.New("slot start time is in the past")
)

func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidRoomID) ||
		errors.Is(err, ErrEmptyDaysOfWeek) ||
		errors.Is(err, ErrInvalidDayOfWeek) ||
		errors.Is(err, ErrDuplicateDayOfWeek) ||
		errors.Is(err, ErrInvalidTimeFormat) ||
		errors.Is(err, ErrEndTimeBeforeStart) ||
		errors.Is(err, ErrSlotTooShort)
}
