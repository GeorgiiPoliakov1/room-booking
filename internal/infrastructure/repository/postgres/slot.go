package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"room-booking/internal/domain"
)

type SlotRepository struct {
	pool *pgxpool.Pool
}

func NewSlotRepository(pool *pgxpool.Pool) *SlotRepository {
	return &SlotRepository{pool: pool}
}

func (r *SlotRepository) GetByID(ctx context.Context, id string) (*domain.Slot, error) {
	const query = `
		SELECT id, room_id, start_time, end_time
		FROM slots
		WHERE id = $1
	`

	var slotID, roomID string
	var startTime, endTime time.Time

	err := r.pool.QueryRow(ctx, query, id).Scan(&slotID, &roomID, &startTime, &endTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repo: get slot by ID: %w", err)
	}

	return domain.NewSlot(slotID, roomID, startTime, endTime)
}

func (r *SlotRepository) GetByRoomAndDate(ctx context.Context, roomID, dateStr string) ([]*domain.Slot, error) {
	const query = `
		SELECT id, room_id, start_time, end_time
		FROM slots
		WHERE room_id = $1 
		  AND start_time >= $2::date 
		  AND start_time < ($2::date + INTERVAL '1 day')
		ORDER BY start_time ASC
	`

	rows, err := r.pool.Query(ctx, query, roomID, dateStr)
	if err != nil {
		return nil, fmt.Errorf("repo: get slots by room and date: %w", err)
	}
	defer rows.Close()

	var slots []*domain.Slot
	for rows.Next() {
		var id, roomID string
		var start, end time.Time
		if err := rows.Scan(&id, &roomID, &start, &end); err != nil {
			return nil, fmt.Errorf("repo: scan slot row: %w", err)
		}
		slot, err := domain.NewSlot(id, roomID, start, end)
		if err != nil {
			return nil, fmt.Errorf("repo: restore slot domain object: %w", err)
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repo: iterate slot rows: %w", err)
	}

	return slots, nil
}

func (r *SlotRepository) Create(ctx context.Context, slot *domain.Slot) error {
	const query = `
        INSERT INTO slots (id, room_id, start_time, end_time)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (room_id, start_time) DO NOTHING
    `
	_, err := r.pool.Exec(ctx, query, slot.ID(), slot.RoomID(), slot.Start(), slot.End())
	if err != nil {
		return fmt.Errorf("repo: create slot: %w", err)
	}
	return nil
}
