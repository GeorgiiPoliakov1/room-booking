package response

import "time"

type Booking struct {
	ID             string     `json:"id" validate:"required,uuid_custom"`
	SlotID         string     `json:"slotId" validate:"required,uuid_custom"`
	UserID         string     `json:"userId" validate:"required,uuid_custom"`
	Status         string     `json:"status" validate:"required,oneof=active cancelled"`
	ConferenceLink *string    `json:"conferenceLink,omitempty"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
}

type BookingCreateResponse struct {
	Booking Booking `json:"booking"`
}
