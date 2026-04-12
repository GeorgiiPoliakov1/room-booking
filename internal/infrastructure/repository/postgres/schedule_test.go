package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"room-booking/internal/domain"
)

func newTestSchedule(t *testing.T, roomID string, days []int, start, end string) *domain.Schedule {
	t.Helper()
	params := domain.ScheduleCreateParams{
		ID:         uuid.New().String(),
		RoomID:     roomID,
		DaysOfWeek: days,
		StartTime:  start,
		EndTime:    end,
	}
	s, err := domain.NewSchedule(params)
	require.NoError(t, err, "domain validation failed in test setup")
	return s
}

func TestScheduleRepository_RoundTrip(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewScheduleRepository(pool)

	roomID := uuid.New().String()
	original := newTestSchedule(t, roomID, []int{1, 3, 5, 7}, "08:30", "10:00")

	require.NoError(t, repo.Create(ctx, original), "Create should succeed")

	found, err := repo.GetByRoomID(ctx, roomID)
	require.NoError(t, err, "GetByRoomID should not return error")
	require.NotNil(t, found, "Schedule should be found")

	require.Equal(t, original.ID(), found.ID())
	require.Equal(t, original.RoomID(), found.RoomID())
	require.Equal(t, original.DaysOfWeek(), found.DaysOfWeek())
	require.Equal(t, original.StartTimeUTC(), found.StartTimeUTC())
	require.Equal(t, original.EndTimeUTC(), found.EndTimeUTC())
	require.True(t, found.CreatedAt().After(time.Now().Add(-15*time.Second)),
		"CreatedAt should be recent and in UTC")
}

func TestScheduleRepository_CreateDuplicate(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewScheduleRepository(pool)

	roomID := uuid.New().String()

	s1 := newTestSchedule(t, roomID, []int{1}, "09:00", "10:00")
	require.NoError(t, repo.Create(ctx, s1))

	s2 := newTestSchedule(t, roomID, []int{2}, "14:00", "15:00")
	err := repo.Create(ctx, s2)

	require.ErrorIs(t, err, domain.ErrScheduleAlreadyExists,
		"Duplicate room_id must return ErrScheduleAlreadyExists")
}

func TestScheduleRepository_GetNotFound(t *testing.T) {
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewScheduleRepository(pool)

	found, err := repo.GetByRoomID(ctx, uuid.New().String())

	require.NoError(t, err, "Missing record must not return DB error")
	require.Nil(t, found, "Missing record must return nil schedule")
}
