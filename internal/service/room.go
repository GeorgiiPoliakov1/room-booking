package service

import (
	"context"
	"errors"

	"room-booking/internal/db/repository"
	"room-booking/internal/domain"

	"github.com/google/uuid"
)

type RoomService interface {
	CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error)
	GetRoom(ctx context.Context, id uuid.UUID) (*domain.Room, error)
	UpdateRoom(ctx context.Context, id uuid.UUID, name string, description *string, capacity *int) (*domain.Room, error)
	ListRooms(ctx context.Context, limit, offset int) ([]*domain.Room, error)

	IsRoomAvailable(ctx context.Context, id uuid.UUID, guests int) (bool, error)
}

type roomService struct {
	repo repository.RoomRepository
}

func (s *roomService) CreateRoom(ctx context.Context, name string, description *string, capacity *int) (*domain.Room, error) {
	room, err := domain.NewRoom(name, description, capacity)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetByName(ctx, name)
	if err == nil && existing != nil {
		return nil, domain.ErrRoomNameExists
	}
	if err != nil && !errors.Is(err, domain.ErrRoomNotFound) {
		return nil, err
	}

	if err := s.repo.Create(ctx, room); err != nil {
		if errors.Is(err, domain.ErrRoomNameExists) {
			return nil, domain.ErrRoomNameExists
		}
		return nil, err
	}

	return room, nil
}

func (s *roomService) GetRoom(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *roomService) UpdateRoom(ctx context.Context, id uuid.UUID, name string, description *string, capacity *int) (*domain.Room, error) {

	room, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tempRoom := &domain.Room{
		ID:          id,
		Name:        name,
		Description: description,
		Capacity:    capacity,
	}
	if err := tempRoom.Validate(); err != nil {
		return nil, err
	}
	room.Name = name
	room.Description = description
	room.Capacity = capacity
	room.UpdatedAt = domain.NowUTC()

	if err := s.repo.Update(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

func (s *roomService) ListRooms(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	if limit > 100 {
		limit = 100
	}
	if limit < 1 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *roomService) IsRoomAvailable(ctx context.Context, id uuid.UUID, guests int) (bool, error) {
	room, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return false, err
	}

	if room.Capacity != nil && guests > *room.Capacity {
		return false, nil
	}
	return true, nil
}

func NewRoomService(repo repository.RoomRepository) RoomService {
	return &roomService{
		repo: repo,
	}
}
