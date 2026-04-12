package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"room-booking/internal/domain"
	"room-booking/internal/infrastructure/repository"

	"github.com/google/uuid"
)

type BookingCreateInput struct {
	SlotID               string
	UserID               string
	CreateConferenceLink bool
}

type BookingCreateService struct {
	slotRepo    repository.SlotRepository
	bookingRepo repository.BookingRepository
	nowFunc     func() time.Time
}

func NewBookingCreateService(slotRepo repository.SlotRepository, bookingRepo repository.BookingRepository) *BookingCreateService {
	return &BookingCreateService{
		slotRepo:    slotRepo,
		bookingRepo: bookingRepo,
		nowFunc:     time.Now,
	}
}

func (s *BookingCreateService) Create(ctx context.Context, input BookingCreateInput) (*domain.Booking, error) {
	if _, err := uuid.Parse(input.SlotID); err != nil {
		return nil, domain.ErrInvalidSlotID
	}
	if _, err := uuid.Parse(input.UserID); err != nil {
		return nil, domain.ErrInvalidUserID
	}

	slot, err := s.slotRepo.GetByID(ctx, input.SlotID)
	if err != nil {
		return nil, fmt.Errorf("get slot: %w", err)
	}
	if slot == nil {
		return nil, domain.ErrSlotNotFound
	}

	now := s.nowFunc().UTC()
	if slot.Start().Before(now) {
		return nil, domain.ErrSlotInPast
	}

	var confLink *string
	if input.CreateConferenceLink {
		link := fmt.Sprintf("https://meet.room-booking.com/%s", uuid.New().String())
		confLink = &link
	}

	booking, err := domain.NewBooking(domain.BookingCreateParams{
		ID:             uuid.New().String(),
		SlotID:         input.SlotID,
		UserID:         input.UserID,
		ConferenceLink: confLink,
	})
	if err != nil {
		return nil, fmt.Errorf("create booking domain object: %w", err)
	}

	if err := s.bookingRepo.Create(ctx, booking); err != nil {
		if errors.Is(err, domain.ErrSlotAlreadyBooked) {
			return nil, err
		}
		return nil, fmt.Errorf("persist booking: %w", err)
	}

	return booking, nil
}
