package domain

import (
	"time"
)

type BookingStatus string

const (
	StatusActive    BookingStatus = "active"
	StatusCancelled BookingStatus = "cancelled"
)

func (s BookingStatus) IsValid() bool {
	return s == StatusActive || s == StatusCancelled
}

type Booking struct {
	id             string
	slotID         string
	userID         string
	status         BookingStatus
	conferenceLink *string
	createdAt      time.Time
}

type BookingCreateParams struct {
	ID             string
	SlotID         string
	UserID         string
	ConferenceLink *string
}

func NewBooking(params BookingCreateParams) (*Booking, error) {
	if params.ID == "" {
		return nil, ErrInvalidBookingID
	}
	if params.SlotID == "" {
		return nil, ErrInvalidSlotID
	}
	if params.UserID == "" {
		return nil, ErrInvalidUserID
	}

	return &Booking{
		id:             params.ID,
		slotID:         params.SlotID,
		userID:         params.UserID,
		status:         StatusActive,
		conferenceLink: params.ConferenceLink,
		createdAt:      time.Now().UTC(),
	}, nil
}

func RestoreBooking(id, slotID, userID string, status BookingStatus, confLink *string, createdAt time.Time) (*Booking, error) {
	if !status.IsValid() {
		return nil, ErrInvalidBookingStatus
	}
	return &Booking{
		id:             id,
		slotID:         slotID,
		userID:         userID,
		status:         status,
		conferenceLink: confLink,
		createdAt:      createdAt.UTC(),
	}, nil
}

func (b *Booking) ID() string              { return b.id }
func (b *Booking) SlotID() string          { return b.slotID }
func (b *Booking) UserID() string          { return b.userID }
func (b *Booking) Status() BookingStatus   { return b.status }
func (b *Booking) ConferenceLink() *string { return b.conferenceLink }
func (b *Booking) CreatedAt() time.Time    { return b.createdAt }

func (b *Booking) IsActive() bool {
	return b.status == StatusActive
}

func (b *Booking) Cancel() error {
	if b.status == StatusCancelled {
		return ErrBookingAlreadyCancelled
	}
	b.status = StatusCancelled
	return nil
}
