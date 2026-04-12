package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	MaxRoomNameLength    = 255
	MaxDescriptionLength = 1000
	MinCapacity          = 1
	MaxCapacity          = 100
)

type Room struct {
	ID          uuid.UUID
	Name        string
	Description *string
	Capacity    *int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewRoom(name string, description *string, capacity *int) (*Room, error) {
	if err := validateRoomParams(name, description, capacity); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Room{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Capacity:    capacity,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func validateRoomParams(name string, description *string, capacity *int) error {
	if name == "" {
		return ErrRoomNameRequired
	}
	if len(name) > MaxRoomNameLength {
		return ErrRoomNameTooLong
	}
	if description != nil && len(*description) > MaxDescriptionLength {
		return ErrDescriptionTooLong
	}
	if capacity != nil && (*capacity < MinCapacity || *capacity > MaxCapacity) {
		return ErrCapacityOutOfRange
	}
	return nil
}

func (r *Room) Validate() error {
	if r == nil {
		return errors.New("room is nil")
	}
	if r.ID == uuid.Nil {
		return ErrInvalidRoomID
	}
	return validateRoomParams(r.Name, r.Description, r.Capacity)
}

func (r *Room) Update(name string, description *string, capacity *int) error {
	if err := validateRoomParams(name, description, capacity); err != nil {
		return err
	}
	r.Name = name
	r.Description = description
	r.Capacity = capacity
	r.UpdatedAt = time.Now().UTC()
	return nil
}
