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

type MockRoomRepository struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.Room, error)
	mock.Mock
}

func (m *MockRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) GetByName(ctx context.Context, name string) (*domain.Room, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Room), args.Error(1)
}

func (m *MockRoomRepository) Update(ctx context.Context, room *domain.Room) error {
	args := m.Called(ctx, room)
	return args.Error(0)
}

func (m *MockRoomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoomRepository) List(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Room), args.Error(1)
}

func TestRoomService_CreateRoom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupMock func(*MockRoomRepository)
		inputName string
		inputDesc *string
		inputCap  *int
		wantErr   error
		wantRoom  bool
	}{
		{
			name: "success",
			setupMock: func(m *MockRoomRepository) {
				m.On("GetByName", mock.Anything, "New Room").Return(nil, domain.ErrRoomNotFound)
				m.On("Create", mock.Anything, mock.MatchedBy(func(r *domain.Room) bool {
					return r.Name == "New Room"
				})).Return(nil)
			},
			inputName: "New Room",
			wantErr:   nil,
			wantRoom:  true,
		},
		{
			name: "name already exists",
			setupMock: func(m *MockRoomRepository) {
				m.On("GetByName", mock.Anything, "Existing").Return(&domain.Room{Name: "Existing"}, nil)
			},
			inputName: "Existing",
			wantErr:   domain.ErrRoomNameExists,
			wantRoom:  false,
		},
		{
			name:      "validation error",
			inputName: "",
			wantErr:   domain.ErrRoomNameRequired,
			wantRoom:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := new(MockRoomRepository)
			if tt.setupMock != nil {
				tt.setupMock(mockRepo)
			}

			svc := NewRoomService(mockRepo)
			room, err := svc.CreateRoom(context.Background(), tt.inputName, tt.inputDesc, tt.inputCap)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, room)
				return
			}

			require.NoError(t, err)
			if tt.wantRoom {
				require.NotNil(t, room)
				assert.Equal(t, tt.inputName, room.Name)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRoomService_IsRoomAvailable(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockRoomRepository)
	capacity := 10
	mockRoom := &domain.Room{
		ID:       uuid.New(),
		Name:     "Test Room",
		Capacity: &capacity,
	}

	mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(mockRoom, nil)

	svc := NewRoomService(mockRepo)

	available, err := svc.IsRoomAvailable(context.Background(), mockRoom.ID, 5)
	require.NoError(t, err)
	assert.True(t, available)

	available, err = svc.IsRoomAvailable(context.Background(), mockRoom.ID, 15)
	require.NoError(t, err)
	assert.False(t, available)

	mockRepo.AssertExpectations(t)
}
