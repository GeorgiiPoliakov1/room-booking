package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertTestRoom(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	uniqueName := fmt.Sprintf("Test Room %s", id)
	_, err := pool.Exec(context.Background(), `
		INSERT INTO rooms (id, name, capacity, created_at, updated_at)
		VALUES ($1, $2, 10, NOW(), NOW())
	`, id, uniqueName)
	require.NoError(t, err, "failed to insert test room")
}

func insertTestSlot(t *testing.T, pool *pgxpool.Pool, id, roomID string, start, end time.Time) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO slots (id, room_id, start_time, end_time)
		VALUES ($1, $2, $3, $4)
	`, id, roomID, start, end)
	require.NoError(t, err, "failed to insert test slot")
}

func TestSlotPGRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewSlotRepository(pool)

	roomID := uuid.New().String()
	insertTestRoom(t, pool, roomID)

	slotID := uuid.New().String()
	start := time.Date(2024, 6, 10, 9, 0, 0, 0, time.UTC)
	end := time.Date(2024, 6, 10, 9, 30, 0, 0, time.UTC)
	insertTestSlot(t, pool, slotID, roomID, start, end)

	t.Run("success: found", func(t *testing.T) {
		t.Parallel()
		slot, err := repo.GetByID(ctx, slotID)
		require.NoError(t, err)
		require.NotNil(t, slot)

		assert.Equal(t, slotID, slot.ID())
		assert.Equal(t, roomID, slot.RoomID())
		assert.True(t, slot.Start().Equal(start))
		assert.True(t, slot.End().Equal(end))
	})

	t.Run("not found: returns nil, nil", func(t *testing.T) {
		t.Parallel()
		slot, err := repo.GetByID(ctx, uuid.New().String())
		require.NoError(t, err)
		assert.Nil(t, slot)
	})
}

func TestSlotPGRepo_GetByRoomAndDate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewSlotRepository(pool)

	roomID := uuid.New().String()
	insertTestRoom(t, pool, roomID)

	otherRoomID := uuid.New().String()
	insertTestRoom(t, pool, otherRoomID)

	dateStr := "2024-06-10"
	s1Start := time.Date(2024, 6, 10, 9, 0, 0, 0, time.UTC)
	s1End := time.Date(2024, 6, 10, 9, 30, 0, 0, time.UTC)
	s2Start := time.Date(2024, 6, 10, 10, 0, 0, 0, time.UTC)
	s2End := time.Date(2024, 6, 10, 10, 30, 0, 0, time.UTC)

	insertTestSlot(t, pool, uuid.New().String(), roomID, s1Start, s1End)
	insertTestSlot(t, pool, uuid.New().String(), roomID, s2Start, s2End)
	insertTestSlot(t, pool, uuid.New().String(), otherRoomID, s1Start, s1End)

	t.Run("success: returns slots for specific room and date", func(t *testing.T) {
		t.Parallel()
		slots, err := repo.GetByRoomAndDate(ctx, roomID, dateStr)
		require.NoError(t, err)
		require.Len(t, slots, 2)

		assert.True(t, slots[0].Start().Before(slots[1].Start()), "slots should be ordered by start time")
	})

	t.Run("empty: no slots for given date", func(t *testing.T) {
		t.Parallel()
		slots, err := repo.GetByRoomAndDate(ctx, roomID, "2024-06-15")
		require.NoError(t, err)
		assert.Empty(t, slots)
	})
}
