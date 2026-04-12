package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"room-booking/internal/domain"
	"room-booking/internal/infrastructure/repository"
)

type CreateScheduleInput struct {
	RoomID     string
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

type ScheduleCreateService struct {
	roomRepo     repository.RoomRepository
	scheduleRepo repository.ScheduleRepository
	slotRepo     repository.SlotRepository
}

func NewScheduleService(roomRepo repository.RoomRepository, scheduleRepo repository.ScheduleRepository, slotRepo repository.SlotRepository) *ScheduleCreateService {
	return &ScheduleCreateService{
		roomRepo:     roomRepo,
		scheduleRepo: scheduleRepo,
		slotRepo:     slotRepo,
	}
}

func (s *ScheduleCreateService) Create(ctx context.Context, input CreateScheduleInput) (*domain.Schedule, error) {
	roomUUID, err := uuid.Parse(input.RoomID)
	if err != nil {
		return nil, domain.ErrInvalidRoomID
	}

	room, err := s.roomRepo.GetByID(ctx, roomUUID)
	if err != nil {
		return nil, fmt.Errorf("check room existence: %w", err)
	}
	if room == nil {
		return nil, domain.ErrRoomNotFound
	}

	existing, err := s.scheduleRepo.GetByRoomID(ctx, input.RoomID)
	if err != nil {
		return nil, fmt.Errorf("check existing schedule: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrScheduleAlreadyExists
	}
	schedule, err := domain.NewSchedule(domain.ScheduleCreateParams{
		ID:         uuid.New().String(),
		RoomID:     input.RoomID,
		DaysOfWeek: input.DaysOfWeek,
		StartTime:  input.StartTime,
		EndTime:    input.EndTime,
	})
	if err != nil {
		return nil, err
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, fmt.Errorf("persist schedule: %w", err)
	}

	return schedule, nil
}

func (s *ScheduleCreateService) generateAndPersistSlots(ctx context.Context, schedule *domain.Schedule) error {
	horizon := time.Now().UTC().AddDate(0, 3, 0)
	current := schedule.CreatedAt().UTC()

	for current.Before(horizon) {
		dow := int(current.Weekday())
		if dow == 0 {
			dow = 7
		}

		if containsInt(schedule.DaysOfWeek(), dow) {
			startH, startM := parseHHMM(schedule.StartTimeUTC())
			endH, endM := parseHHMM(schedule.EndTimeUTC())
			dayStart := time.Date(current.Year(), current.Month(), current.Day(), startH, startM, 0, 0, time.UTC)
			dayEnd := time.Date(current.Year(), current.Month(), current.Day(), endH, endM, 0, 0, time.UTC)

			for t := dayStart; t.Add(30 * time.Minute).Before(dayEnd); t = t.Add(30 * time.Minute) {
				slotEnd := t.Add(30 * time.Minute)
				slot, _ := domain.NewSlot(uuid.New().String(), schedule.RoomID(), t, slotEnd)
				if err := s.slotRepo.Create(ctx, slot); err != nil {
					return fmt.Errorf("persist slot: %w", err)
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	return nil
}
