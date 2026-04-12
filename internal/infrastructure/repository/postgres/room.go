package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"room-booking/internal/domain"
)

type RoomRepository struct {
	pool *pgxpool.Pool
}

// NewRoomRepository создаёт экземпляр репозитория для работы с таблицей rooms.
func NewRoomRepository(pool *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	const query = `
		INSERT INTO rooms (id, name, description, capacity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		room.ID, room.Name, room.Description, room.Capacity,
		room.CreatedAt, room.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrRoomNameExists
		}
		return fmt.Errorf("repo: create room: %w", err)
	}
	return nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Room, error) {
	const query = `
		SELECT id, name, description, capacity, created_at, updated_at
		FROM rooms
		WHERE id = $1
	`

	var room domain.Room
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&room.ID, &room.Name, &room.Description, &room.Capacity,
		&room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("repo: get room by ID: %w", err)
	}
	return &room, nil
}

func (r *RoomRepository) Update(ctx context.Context, room *domain.Room) error {
	const query = `
		UPDATE rooms 
		SET name = $1, description = $2, capacity = $3, updated_at = $4
		WHERE id = $5
	`

	tag, err := r.pool.Exec(ctx, query,
		room.Name, room.Description, room.Capacity, room.UpdatedAt, room.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrRoomNameExists
		}
		return fmt.Errorf("repo: update room: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (r *RoomRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM rooms WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("repo: delete room: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (r *RoomRepository) List(ctx context.Context, limit, offset int) ([]*domain.Room, error) {
	const query = `
		SELECT id, name, description, capacity, created_at, updated_at
		FROM rooms
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repo: list rooms: %w", err)
	}
	defer rows.Close()

	rooms, err := pgx.CollectRows(rows, pgx.RowToStructByPos[domain.Room])
	if err != nil {
		return nil, fmt.Errorf("repo: collect rooms: %w", err)
	}

	// Конвертируем []domain.Room в []*domain.Room
	result := make([]*domain.Room, len(rooms))
	for i := range rooms {
		result[i] = &rooms[i]
	}
	return result, nil
}

func (r *RoomRepository) GetByName(ctx context.Context, name string) (*domain.Room, error) {
	const query = `
		SELECT id, name, description, capacity, created_at, updated_at
		FROM rooms
		WHERE name = $1
		LIMIT 1
	`

	var room domain.Room
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&room.ID, &room.Name, &room.Description, &room.Capacity,
		&room.CreatedAt, &room.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrRoomNotFound
		}
		return nil, fmt.Errorf("repo: get room by name: %w", err)
	}
	return &room, nil
}
