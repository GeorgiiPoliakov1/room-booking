package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(i int) *int       { return &i }
func strPtr(s string) *string { return &s }

func TestNewRoom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		roomName    string
		description *string
		capacity    *int
		wantErr     error
		wantRoom    bool
	}{
		{
			name:        "valid minimal room",
			roomName:    "Conference A",
			description: nil,
			capacity:    nil,
			wantErr:     nil,
			wantRoom:    true,
		},
		{
			name:        "valid with all fields",
			roomName:    "Board Room",
			description: strPtr("Executive meeting room"),
			capacity:    intPtr(12),
			wantErr:     nil,
			wantRoom:    true,
		},
		{
			name:        "empty name",
			roomName:    "",
			description: nil,
			capacity:    nil,
			wantErr:     ErrRoomNameRequired,
			wantRoom:    false,
		},
		{
			name:        "name too long",
			roomName:    strings.Repeat("x", MaxRoomNameLength+1),
			description: nil,
			capacity:    nil,
			wantErr:     ErrRoomNameTooLong,
			wantRoom:    false,
		},
		{
			name:        "description too long",
			roomName:    "Room",
			description: strPtr(strings.Repeat("y", MaxDescriptionLength+1)),
			capacity:    nil,
			wantErr:     ErrDescriptionTooLong,
			wantRoom:    false,
		},
		{
			name:        "capacity below min",
			roomName:    "Room",
			description: nil,
			capacity:    intPtr(0),
			wantErr:     ErrCapacityOutOfRange,
			wantRoom:    false,
		},
		{
			name:        "capacity above max",
			roomName:    "Room",
			description: nil,
			capacity:    intPtr(101),
			wantErr:     ErrCapacityOutOfRange,
			wantRoom:    false,
		},
		{
			name:        "boundary: capacity = 1",
			roomName:    "Room",
			description: nil,
			capacity:    intPtr(1),
			wantErr:     nil,
			wantRoom:    true,
		},
		{
			name:        "boundary: capacity = 100",
			roomName:    "Room",
			description: nil,
			capacity:    intPtr(100),
			wantErr:     nil,
			wantRoom:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			room, err := NewRoom(tt.roomName, tt.description, tt.capacity)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr, "expected specific error")
				assert.Nil(t, room, "room should be nil on error")
				return
			}

			require.NoError(t, err, "unexpected error")
			require.NotNil(t, room, "room should not be nil")

			assert.NotEqual(t, uuid.Nil, room.ID, "ID should be generated")
			assert.Equal(t, tt.roomName, room.Name)
			assert.Equal(t, tt.description, room.Description)
			assert.Equal(t, tt.capacity, room.Capacity)
			assert.WithinDuration(t, time.Now().UTC(), room.CreatedAt, 2*time.Second)
			assert.Equal(t, room.CreatedAt, room.UpdatedAt, "Created and Updated should match on creation")
		})
	}
}

func TestRoom_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		room    *Room
		wantErr error
	}{
		{
			name: "valid room",
			room: &Room{
				ID:   uuid.New(),
				Name: "Valid Room",
			},
			wantErr: nil,
		},
		{
			name:    "nil room",
			room:    nil,
			wantErr: errors.New("room is nil"),
		},
		{
			name: "nil ID",
			room: &Room{
				ID:   uuid.Nil,
				Name: "Bad Room",
			},
			wantErr: ErrInvalidRoomID,
		},
		{
			name: "empty name",
			room: &Room{
				ID:   uuid.New(),
				Name: "",
			},
			wantErr: ErrRoomNameRequired,
		},
		{
			name: "negative capacity",
			room: &Room{
				ID:       uuid.New(),
				Name:     "Room",
				Capacity: intPtr(-5),
			},
			wantErr: ErrCapacityOutOfRange,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.room.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRoom_Update(t *testing.T) {
	t.Parallel()

	room, err := NewRoom("Original", strPtr("desc"), intPtr(10))
	require.NoError(t, err)
	require.NotNil(t, room)

	oldUpdatedAt := room.UpdatedAt
	time.Sleep(10 * time.Millisecond)

	err = room.Update("Updated Name", strPtr("new desc"), intPtr(20))
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", room.Name)
	assert.Equal(t, "new desc", *room.Description)
	assert.Equal(t, 20, *room.Capacity)
	assert.True(t, room.UpdatedAt.After(oldUpdatedAt), "UpdatedAt should be refreshed")
}

func TestRoom_Update_InvalidData(t *testing.T) {
	t.Parallel()

	room, err := NewRoom("Original", nil, nil)
	require.NoError(t, err)

	err = room.Update("", nil, nil)
	assert.ErrorIs(t, err, ErrRoomNameRequired)

	assert.Equal(t, "Original", room.Name)
}
