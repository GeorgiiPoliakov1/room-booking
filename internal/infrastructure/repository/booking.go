package repository

import (
	"context"
	"room-booking/internal/domain"
)

type BookingRepository interface {
	Create(ctx context.Context, booking *domain.Booking) error
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	Cancel(ctx context.Context, id string) error
}
