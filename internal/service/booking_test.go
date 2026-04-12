package service

import (
	"context"
	"testing"
	"time"

	"room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSlotRepository struct {
	mock.Mock
}

func (m *MockSlotRepository) GetByID(ctx context.Context, id string) (*domain.Slot, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Slot), args.Error(1)
}

func (m *MockSlotRepository) GetByRoomAndDate(ctx context.Context, roomID, date string) ([]*domain.Slot, error) {
	args := m.Called(ctx, roomID, date)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Slot), args.Error(1)
}
func (m *MockSlotRepository) Create(ctx context.Context, slot *domain.Slot) error {
	args := m.Called(ctx, slot)
	return args.Error(0)
}

type MockBookingRepository struct {
	mock.Mock
}

func (m *MockBookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	args := m.Called(ctx, booking)
	return args.Error(0)
}

func (m *MockBookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Booking), args.Error(1)
}

func (m *MockBookingRepository) Cancel(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestBookingCreateService_Create(t *testing.T) {
	t.Parallel()

	now := time.Date(2024, 6, 10, 12, 0, 0, 0, time.UTC)

	validSlotID := uuid.New().String()
	pastSlotID := uuid.New().String()
	userID := uuid.New().String()

	futureStart := now.Add(2 * time.Hour)
	futureEnd := futureStart.Add(30 * time.Minute)
	validSlot, err := domain.NewSlot(validSlotID, uuid.New().String(), futureStart, futureEnd)
	require.NoError(t, err)

	pastStart := now.Add(-2 * time.Hour)
	pastEnd := pastStart.Add(30 * time.Minute)
	pastSlot, err := domain.NewSlot(pastSlotID, uuid.New().String(), pastStart, pastEnd)
	require.NoError(t, err)

	tests := []struct {
		name        string
		setupMocks  func(*MockSlotRepository, *MockBookingRepository)
		input       BookingCreateInput
		wantErr     error
		checkResult func(t *testing.T, booking *domain.Booking)
	}{
		{
			name: "success: booking without conference link",
			setupMocks: func(msr *MockSlotRepository, mbr *MockBookingRepository) {
				msr.On("GetByID", mock.Anything, validSlotID).Return(validSlot, nil)
				mbr.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Booking) bool {
					return b.ConferenceLink() == nil && b.SlotID() == validSlotID && b.UserID() == userID
				})).Return(nil)
			},
			input: BookingCreateInput{SlotID: validSlotID, UserID: userID, CreateConferenceLink: false},
			checkResult: func(t *testing.T, b *domain.Booking) {
				require.NotNil(t, b)
				assert.Nil(t, b.ConferenceLink())
				assert.Equal(t, domain.StatusActive, b.Status())
			},
		},
		{
			name: "success: booking with conference link",
			setupMocks: func(msr *MockSlotRepository, mbr *MockBookingRepository) {
				msr.On("GetByID", mock.Anything, validSlotID).Return(validSlot, nil)
				mbr.On("Create", mock.Anything, mock.MatchedBy(func(b *domain.Booking) bool {
					return b.ConferenceLink() != nil && b.SlotID() == validSlotID
				})).Return(nil)
			},
			input: BookingCreateInput{SlotID: validSlotID, UserID: userID, CreateConferenceLink: true},
			checkResult: func(t *testing.T, b *domain.Booking) {
				require.NotNil(t, b)
				require.NotNil(t, b.ConferenceLink())
				assert.Contains(t, *b.ConferenceLink(), "https://meet.room-booking.com/")
			},
		},
		{
			name: "404: slot not found",
			setupMocks: func(msr *MockSlotRepository, mbr *MockBookingRepository) {
				msr.On("GetByID", mock.Anything, validSlotID).Return(nil, nil)
			},
			input:   BookingCreateInput{SlotID: validSlotID, UserID: userID},
			wantErr: domain.ErrSlotNotFound,
		},
		{
			name: "400: slot in the past",
			setupMocks: func(msr *MockSlotRepository, mbr *MockBookingRepository) {
				msr.On("GetByID", mock.Anything, pastSlotID).Return(pastSlot, nil)
			},
			input:   BookingCreateInput{SlotID: pastSlotID, UserID: userID},
			wantErr: domain.ErrSlotInPast,
		},
		{
			name:    "400: invalid slot ID format",
			input:   BookingCreateInput{SlotID: "not-a-uuid", UserID: userID},
			wantErr: domain.ErrInvalidSlotID,
		},
		{
			name:    "400: invalid user ID format",
			input:   BookingCreateInput{SlotID: validSlotID, UserID: "not-a-uuid"},
			wantErr: domain.ErrInvalidUserID,
		},
		{
			name: "409: slot already booked",
			setupMocks: func(msr *MockSlotRepository, mbr *MockBookingRepository) {
				msr.On("GetByID", mock.Anything, validSlotID).Return(validSlot, nil)
				mbr.On("Create", mock.Anything, mock.AnythingOfType("*domain.Booking")).
					Return(domain.ErrSlotAlreadyBooked)
			},
			input:   BookingCreateInput{SlotID: validSlotID, UserID: userID},
			wantErr: domain.ErrSlotAlreadyBooked,
		},
		{
			name: "500: repository persistence error",
			setupMocks: func(msr *MockSlotRepository, mbr *MockBookingRepository) {
				msr.On("GetByID", mock.Anything, validSlotID).Return(validSlot, nil)
				mbr.On("Create", mock.Anything, mock.AnythingOfType("*domain.Booking")).
					Return(assert.AnError)
			},
			input:   BookingCreateInput{SlotID: validSlotID, UserID: userID},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			msr := new(MockSlotRepository)
			mbr := new(MockBookingRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(msr, mbr)
			}

			svc := NewBookingCreateService(msr, mbr)
			svc.nowFunc = func() time.Time { return now }

			booking, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, booking)
				msr.AssertExpectations(t)
				mbr.AssertExpectations(t)
				return
			}

			if tt.name == "500: repository persistence error" {
				if err != nil {
					assert.Error(t, err)
					assert.Nil(t, booking)
					msr.AssertExpectations(t)
					mbr.AssertExpectations(t)
					return
				}
			}

			require.NoError(t, err)
			require.NotNil(t, booking)

			if tt.checkResult != nil {
				tt.checkResult(t, booking)
			}

			msr.AssertExpectations(t)
			mbr.AssertExpectations(t)
		})
	}
}
