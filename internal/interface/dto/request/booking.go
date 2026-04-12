package request

type BookingCreateRequest struct {
	SlotID               string `json:"slotId" validate:"required,uuid_custom"`
	CreateConferenceLink *bool  `json:"createConferenceLink,omitempty"`
}
