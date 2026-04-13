package repository

import (
	"context"

	"room-booking/internal/domain"
)

type SlotRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Slot, error)
	GetByRoomAndDate(ctx context.Context, roomID string, date string) ([]*domain.Slot, error)
	Create(ctx context.Context, slot *domain.Slot) error
}
