package service

import (
	"context"
	"fmt"
	"time"

	"room-booking/internal/domain"
	"room-booking/internal/infrastructure/repository"

	"github.com/google/uuid"
)

type SlotListService struct {
	roomRepo repository.RoomRepository
	slotRepo repository.SlotRepository
}

func NewSlotListService(
	roomRepo repository.RoomRepository,
	slotRepo repository.SlotRepository,
) *SlotListService {
	return &SlotListService{
		roomRepo: roomRepo,
		slotRepo: slotRepo,
	}
}

func (s *SlotListService) GetAvailableSlots(ctx context.Context, roomID, dateStr string) ([]domain.Slot, error) {
	roomUUID, err := uuid.Parse(roomID)
	if err != nil {
		return nil, domain.ErrInvalidRoomID
	}
	room, err := s.roomRepo.GetByID(ctx, roomUUID)
	if err != nil {
		return nil, fmt.Errorf("check room: %w", err)
	}
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}

	slotsPtr, err := s.slotRepo.GetByRoomAndDate(ctx, roomID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("get slots from db: %w", err)
	}

	slots := make([]domain.Slot, len(slotsPtr))
	for i, s := range slotsPtr {
		slots[i] = *s
	}
	return slots, nil
}

func containsInt(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func parseHHMM(s string) (h, m int) {
	t, _ := time.Parse("15:04", s)
	return t.Hour(), t.Minute()
}
