package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"room-booking/internal/domain"
)

func makeTestBooking(t *testing.T, slotID, userID string, confLink *string) *domain.Booking {
	t.Helper()
	params := domain.BookingCreateParams{
		ID:             uuid.New().String(),
		SlotID:         slotID,
		UserID:         userID,
		ConferenceLink: confLink,
	}
	b, err := domain.NewBooking(params)
	require.NoError(t, err, "domain.NewBooking failed in test setup")
	return b
}

func TestBookingRepository_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewBookingRepository(pool)

	slotID := uuid.New().String()
	userID := uuid.New().String()

	t.Run("success: with conference link", func(t *testing.T) {
		link := "https://meet.example.com/abc-123"
		booking := makeTestBooking(t, slotID, userID, &link)
		require.NoError(t, repo.Create(ctx, booking))

		got, err := repo.GetByID(ctx, booking.ID())
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, booking.ID(), got.ID())
		assert.Equal(t, domain.StatusActive, got.Status())
		require.NotNil(t, got.ConferenceLink())
		assert.Equal(t, &link, got.ConferenceLink())
	})

	t.Run("success: without conference link (nil)", func(t *testing.T) {
		slotID2 := uuid.New().String()
		booking := makeTestBooking(t, slotID2, userID, nil)
		require.NoError(t, repo.Create(ctx, booking))

		got, err := repo.GetByID(ctx, booking.ID())
		require.NoError(t, err)
		assert.Nil(t, got.ConferenceLink())
		assert.Equal(t, domain.StatusActive, got.Status())
	})

	t.Run("fail: duplicate active booking for same slot (409)", func(t *testing.T) {
		slotID3 := uuid.New().String()
		b1 := makeTestBooking(t, slotID3, userID, nil)
		require.NoError(t, repo.Create(ctx, b1))

		b2 := makeTestBooking(t, slotID3, uuid.New().String(), nil)
		err := repo.Create(ctx, b2)
		require.ErrorIs(t, err, domain.ErrSlotAlreadyBooked)
	})
}

func TestBookingRepository_GetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewBookingRepository(pool)

	slotID := uuid.New().String()
	userID := uuid.New().String()
	original := makeTestBooking(t, slotID, userID, nil)
	require.NoError(t, repo.Create(ctx, original))

	t.Run("success: found", func(t *testing.T) {
		got, err := repo.GetByID(ctx, original.ID())
		require.NoError(t, err)
		require.NotNil(t, got)

		assert.Equal(t, original.ID(), got.ID())
		assert.Equal(t, original.SlotID(), got.SlotID())
		assert.Equal(t, original.UserID(), got.UserID())
		assert.Equal(t, domain.StatusActive, got.Status())
		assert.True(t, got.CreatedAt().After(time.Now().Add(-10*time.Second)),
			"CreatedAt should be recent")
	})

	t.Run("not found: returns nil, nil", func(t *testing.T) {
		got, err := repo.GetByID(ctx, uuid.New().String())
		require.NoError(t, err)
		assert.Nil(t, got)
	})
}

func TestBookingRepository_Cancel(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	pool := setupTestDB(t)
	repo := NewBookingRepository(pool)

	slotID := uuid.New().String()
	userID := uuid.New().String()
	booking := makeTestBooking(t, slotID, userID, nil)
	require.NoError(t, repo.Create(ctx, booking))

	t.Run("success: active → cancelled", func(t *testing.T) {
		err := repo.Cancel(ctx, booking.ID())
		require.NoError(t, err)

		got, err := repo.GetByID(ctx, booking.ID())
		require.NoError(t, err)
		require.Equal(t, domain.StatusCancelled, got.Status())
	})

	t.Run("fail: already cancelled", func(t *testing.T) {
		err := repo.Cancel(ctx, booking.ID())
		require.ErrorIs(t, err, domain.ErrBookingAlreadyCancelled)
	})

	t.Run("fail: not found", func(t *testing.T) {
		err := repo.Cancel(ctx, uuid.New().String())
		require.ErrorIs(t, err, domain.ErrBookingNotFound)
	})
}
