package repository

import (
	"context"

	"room-booking/internal/domain"
)

type ScheduleRepository interface {
	GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error)
	Create(ctx context.Context, schedule *domain.Schedule) error
}
