package request

import (
	"github.com/google/uuid"
)

type RoomCreateRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=255,room_name"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Capacity    *int    `json:"capacity" validate:"omitempty,min=1,max=100"`
}

type RoomUpdateRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=255,room_name"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	Capacity    *int    `json:"capacity" validate:"omitempty,min=1,max=100"`
}

type RoomListRequest struct {
	Page  int `json:"page" validate:"min=1"`
	Limit int `json:"limit" validate:"min=1,max=100"`
}

type RoomScheduleRequest struct {
	RoomID     uuid.UUID `json:"room_id" validate:"required"`
	DaysOfWeek []int     `json:"days_of_week" validate:"required,min=1,max=7,dive,min=1,max=7"`
	StartTime  string    `json:"start_time" validate:"required,datetime=15:04"`
	EndTime    string    `json:"end_time" validate:"required,datetime=15:04,gtfield=StartTime"`
}
