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

func TestSlotListService_GetAvailableSlots(t *testing.T) {
	t.Parallel()

	roomID := uuid.New().String()
	validRoom := &domain.Room{ID: uuid.MustParse(roomID)}
	testDate := "2024-06-10"

	makeSlots := func(roomID string, startTimes []time.Time) []*domain.Slot {
		var slots []*domain.Slot
		for _, st := range startTimes {
			slot, _ := domain.NewSlot(
				uuid.New().String(),
				roomID,
				st,
				st.Add(30*time.Minute),
			)
			slots = append(slots, slot)
		}
		return slots
	}

	tests := []struct {
		name          string
		setupMocks    func(*MockRoomRepository, *MockSlotRepository)
		inputRoomID   string
		inputDate     string
		wantErr       error
		wantSlotCount int
		verifySlots   func(t *testing.T, slots []domain.Slot)
	}{
		{
			name: "success: 2-hour schedule generates 4 slots",
			setupMocks: func(mr *MockRoomRepository, msr *MockSlotRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(validRoom, nil)
				slots := makeSlots(roomID, []time.Time{
					time.Date(2024, 6, 10, 9, 0, 0, 0, time.UTC),
					time.Date(2024, 6, 10, 9, 30, 0, 0, time.UTC),
					time.Date(2024, 6, 10, 10, 0, 0, 0, time.UTC),
					time.Date(2024, 6, 10, 10, 30, 0, 0, time.UTC),
				})
				msr.On("GetByRoomAndDate", mock.Anything, roomID, testDate).
					Return(slots, nil)
			},
			inputRoomID:   roomID,
			inputDate:     testDate,
			wantErr:       nil,
			wantSlotCount: 4,
			verifySlots: func(t *testing.T, slots []domain.Slot) {
				require.Len(t, slots, 4)
				assert.Equal(t, time.Date(2024, 6, 10, 9, 0, 0, 0, time.UTC), slots[0].Start())
				assert.Equal(t, time.Date(2024, 6, 10, 11, 0, 0, 0, time.UTC), slots[3].End())
			},
		},
		{
			name: "success: exact 30-min window generates 1 slot",
			setupMocks: func(mr *MockRoomRepository, msr *MockSlotRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(validRoom, nil)
				slots := makeSlots(roomID, []time.Time{
					time.Date(2024, 6, 10, 14, 0, 0, 0, time.UTC),
				})
				msr.On("GetByRoomAndDate", mock.Anything, roomID, testDate).
					Return(slots, nil)
			},
			inputRoomID:   roomID,
			inputDate:     testDate,
			wantErr:       nil,
			wantSlotCount: 1,
		},
		{
			name: "success: no slots for date → empty list",
			setupMocks: func(mr *MockRoomRepository, msr *MockSlotRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(validRoom, nil)
				msr.On("GetByRoomAndDate", mock.Anything, roomID, testDate).
					Return([]*domain.Slot{}, nil)
			},
			inputRoomID:   roomID,
			inputDate:     testDate,
			wantErr:       nil,
			wantSlotCount: 0,
		},
		{
			name: "fail: room not found",
			setupMocks: func(mr *MockRoomRepository, msr *MockSlotRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, nil)
			},
			inputRoomID:   roomID,
			inputDate:     testDate,
			wantErr:       domain.ErrRoomNotFound,
			wantSlotCount: 0,
		},
		{
			name:          "fail: invalid room ID format",
			setupMocks:    func(mr *MockRoomRepository, msr *MockSlotRepository) {},
			inputRoomID:   "not-a-uuid",
			inputDate:     testDate,
			wantErr:       domain.ErrInvalidRoomID,
			wantSlotCount: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRoomRepo := new(MockRoomRepository)
			mockSlotRepo := new(MockSlotRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRoomRepo, mockSlotRepo)
			}

			svc := NewSlotListService(mockRoomRepo, mockSlotRepo)
			slots, err := svc.GetAvailableSlots(context.Background(), tt.inputRoomID, tt.inputDate)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, slots)
				mockRoomRepo.AssertExpectations(t)
				mockSlotRepo.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			require.Len(t, slots, tt.wantSlotCount)

			if tt.verifySlots != nil {
				tt.verifySlots(t, slots)
			}

			mockRoomRepo.AssertExpectations(t)
			mockSlotRepo.AssertExpectations(t)
		})
	}
}
