package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"room-booking/internal/domain"
)

type BookingRepository struct {
	pool *pgxpool.Pool
}

func NewBookingRepository(pool *pgxpool.Pool) *BookingRepository {
	return &BookingRepository{pool: pool}
}

func (r *BookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	const query = `
		INSERT INTO bookings (id, slot_id, user_id, status, conference_link, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		booking.ID(),
		booking.SlotID(),
		booking.UserID(),
		booking.Status(),
		booking.ConferenceLink(),
		booking.CreatedAt(),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSlotAlreadyBooked
		}
		return fmt.Errorf("repo: create booking: %w", err)
	}

	return nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	const query = `
		SELECT id, slot_id, user_id, status, conference_link, created_at
		FROM bookings
		WHERE id = $1
	`

	var (
		bID, slotID, userID, statusStr string
		confLink                       *string
		createdAt                      time.Time
	)

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&bID, &slotID, &userID, &statusStr, &confLink, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repo: get booking by ID: %w", err)
	}

	return domain.RestoreBooking(bID, slotID, userID, domain.BookingStatus(statusStr), confLink, createdAt)
}

func (r *BookingRepository) Cancel(ctx context.Context, id string) error {
	const query = `
		UPDATE bookings SET status = 'cancelled'
		WHERE id = $1 AND status = 'active'
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repo: cancel booking: %w", err)
	}

	if tag.RowsAffected() == 0 {
		var exists bool
		if err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM bookings WHERE id = $1)", id).Scan(&exists); err != nil {
			return fmt.Errorf("repo: check booking existence: %w", err)
		}

		if !exists {
			return domain.ErrBookingNotFound
		}
		return domain.ErrBookingAlreadyCancelled
	}

	return nil
}
