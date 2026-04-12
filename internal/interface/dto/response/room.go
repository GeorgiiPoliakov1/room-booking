package response

import (
	"time"

	"github.com/google/uuid"
)

type RoomResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Capacity    *int      `json:"capacity,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomsListResponse struct {
	Rooms []RoomResponse `json:"rooms"`
	Total int            `json:"total,omitempty"`
	Page  int            `json:"page,omitempty"`
	Limit int            `json:"limit,omitempty"`
}

type ErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details,omitempty"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
