package response

import "time"

type Slot struct {
	ID     string    `json:"id" validate:"required,uuid_custom"`
	RoomID string    `json:"roomId" validate:"required,uuid_custom"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}
type SlotListResponse struct {
	Slots []Slot `json:"slots"`
}
