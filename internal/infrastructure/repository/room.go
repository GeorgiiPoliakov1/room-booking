package repository

import (
	"context"

	model "room-booking/internal/domain"

	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room *model.Room) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Room, error)
	GetByName(ctx context.Context, name string) (*model.Room, error)
	Update(ctx context.Context, room *model.Room) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*model.Room, error)
}
