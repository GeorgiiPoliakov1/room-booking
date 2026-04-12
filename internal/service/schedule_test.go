package service

import (
	"context"
	"testing"

	"room-booking/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockScheduleRepository struct {
	mock.Mock
}

func (m *MockScheduleRepository) Create(ctx context.Context, schedule *domain.Schedule) error {
	args := m.Called(ctx, schedule)
	return args.Error(0)
}

func (m *MockScheduleRepository) GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error) {
	args := m.Called(ctx, roomID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Schedule), args.Error(1)
}

func TestScheduleCreateService_CreateSchedule(t *testing.T) {
	t.Parallel()

	validRoomID := uuid.New().String()

	tests := []struct {
		name         string
		setupMocks   func(*MockRoomRepository, *MockScheduleRepository)
		input        CreateScheduleInput
		wantErr      error
		wantSchedule bool
	}{
		{
			name: "success: valid schedule",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domain.Room{ID: uuid.MustParse(validRoomID)}, nil)
				ms.On("GetByRoomID", mock.Anything, validRoomID).
					Return(nil, nil)
				ms.On("Create", mock.Anything, mock.MatchedBy(func(s *domain.Schedule) bool {
					return s.RoomID() == validRoomID && len(s.DaysOfWeek()) == 3
				})).Return(nil)
			},
			input: CreateScheduleInput{
				RoomID:     validRoomID,
				DaysOfWeek: []int{1, 3, 5},
				StartTime:  "09:00",
				EndTime:    "10:30",
			},
			wantErr:      nil,
			wantSchedule: true,
		},
		{
			name: "404: room not found",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(nil, nil)
			},
			input: CreateScheduleInput{
				RoomID: validRoomID, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00",
			},
			wantErr:      domain.ErrRoomNotFound,
			wantSchedule: false,
		},
		{
			name: "409: schedule already exists",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domain.Room{ID: uuid.MustParse(validRoomID)}, nil)
				ms.On("GetByRoomID", mock.Anything, validRoomID).
					Return(&domain.Schedule{}, nil)
			},
			input: CreateScheduleInput{
				RoomID: validRoomID, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00",
			},
			wantErr:      domain.ErrScheduleAlreadyExists,
			wantSchedule: false,
		},
		{
			name: "400: slot duration < 30 minutes",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domain.Room{ID: uuid.MustParse(validRoomID)}, nil)
				ms.On("GetByRoomID", mock.Anything, validRoomID).
					Return(nil, nil)
			},
			input: CreateScheduleInput{
				RoomID: validRoomID, DaysOfWeek: []int{1}, StartTime: "10:00", EndTime: "10:20",
			},
			wantErr:      domain.ErrSlotTooShort,
			wantSchedule: false,
		},
		{
			name: "400: duplicate days in week",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domain.Room{ID: uuid.MustParse(validRoomID)}, nil)
				ms.On("GetByRoomID", mock.Anything, validRoomID).
					Return(nil, nil)
			},
			input: CreateScheduleInput{
				RoomID: validRoomID, DaysOfWeek: []int{1, 1, 3}, StartTime: "09:00", EndTime: "10:00",
			},
			wantErr:      domain.ErrDuplicateDayOfWeek,
			wantSchedule: false,
		},
		{
			name: "400: invalid day of week (0)",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domain.Room{ID: uuid.MustParse(validRoomID)}, nil)
				ms.On("GetByRoomID", mock.Anything, validRoomID).
					Return(nil, nil)
			},
			input: CreateScheduleInput{
				RoomID: validRoomID, DaysOfWeek: []int{0}, StartTime: "09:00", EndTime: "10:00",
			},
			wantErr:      domain.ErrInvalidDayOfWeek,
			wantSchedule: false,
		},
		{
			name: "400: end time before start time",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domain.Room{ID: uuid.MustParse(validRoomID)}, nil)
				ms.On("GetByRoomID", mock.Anything, validRoomID).
					Return(nil, nil)
			},
			input: CreateScheduleInput{
				RoomID: validRoomID, DaysOfWeek: []int{1}, StartTime: "10:00", EndTime: "09:00",
			},
			wantErr:      domain.ErrEndTimeBeforeStart,
			wantSchedule: false,
		},
		{
			name: "500: repository error on create",
			setupMocks: func(mr *MockRoomRepository, ms *MockScheduleRepository) {
				mr.On("GetByID", mock.Anything, mock.AnythingOfType("uuid.UUID")).
					Return(&domain.Room{ID: uuid.MustParse(validRoomID)}, nil)
				ms.On("GetByRoomID", mock.Anything, validRoomID).
					Return(nil, nil)
				ms.On("Create", mock.Anything, mock.AnythingOfType("*domain.Schedule")).
					Return(assert.AnError)
			},
			input: CreateScheduleInput{
				RoomID: validRoomID, DaysOfWeek: []int{1}, StartTime: "09:00", EndTime: "10:00",
			},
			wantErr:      nil,
			wantSchedule: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRoomRepo := new(MockRoomRepository)
			mockScheduleRepo := new(MockScheduleRepository)
			mockSlotRepo := new(MockSlotRepository)

			if tt.setupMocks != nil {
				tt.setupMocks(mockRoomRepo, mockScheduleRepo)
			}

			svc := NewScheduleService(mockRoomRepo, mockScheduleRepo, mockSlotRepo)
			schedule, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, schedule)
				mockRoomRepo.AssertExpectations(t)
				mockScheduleRepo.AssertExpectations(t)
				return
			}

			if !tt.wantSchedule {
				assert.Error(t, err)
				assert.Nil(t, schedule)
				mockRoomRepo.AssertExpectations(t)
				mockScheduleRepo.AssertExpectations(t)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, schedule)
			assert.Equal(t, tt.input.RoomID, schedule.RoomID())
			assert.Equal(t, tt.input.DaysOfWeek, schedule.DaysOfWeek())
			assert.Equal(t, tt.input.StartTime, schedule.StartTimeUTC())
			assert.Equal(t, tt.input.EndTime, schedule.EndTimeUTC())

			mockRoomRepo.AssertExpectations(t)
			mockScheduleRepo.AssertExpectations(t)
		})
	}
}
