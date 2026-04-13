package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"room-booking/internal/db/repository"
	"room-booking/internal/domain"
)

type schedulePGRepo struct {
	pool *pgxpool.Pool
}

func NewScheduleRepository(pool *pgxpool.Pool) repository.ScheduleRepository {
	return &schedulePGRepo{pool: pool}
}

func (r *schedulePGRepo) Create(ctx context.Context, schedule *domain.Schedule) error {
	const query = `
		INSERT INTO schedules (id, room_id, days_of_week, start_time, end_time, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	daysInt32 := make([]int32, len(schedule.DaysOfWeek()))
	for i, d := range schedule.DaysOfWeek() {
		daysInt32[i] = int32(d)
	}

	startT, _ := time.Parse("15:04", schedule.StartTimeUTC())
	endT, _ := time.Parse("15:04", schedule.EndTimeUTC())

	_, err := r.pool.Exec(ctx, query,
		schedule.ID(),
		schedule.RoomID(),
		daysInt32,
		startT,
		endT,
		schedule.CreatedAt(),
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrScheduleAlreadyExists
		}
		return fmt.Errorf("repo: create schedule: %w", err)
	}

	return nil
}

func (r *schedulePGRepo) GetByRoomID(ctx context.Context, roomID string) (*domain.Schedule, error) {
	const query = `
		SELECT id, room_id, days_of_week, start_time, end_time, created_at 
		FROM schedules 
		WHERE room_id = $1
	`

	var (
		idDB, roomDB string
		daysOfWeek   []int32
		startTime    time.Time
		endTime      time.Time
		createdAt    time.Time
	)

	err := r.pool.QueryRow(ctx, query, roomID).Scan(
		&idDB, &roomDB, &daysOfWeek, &startTime, &endTime, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("repo: get schedule by roomID: %w", err)
	}

	days := make([]int, len(daysOfWeek))
	for i, d := range daysOfWeek {
		days[i] = int(d)
	}

	startMins := startTime.Hour()*60 + startTime.Minute()
	endMins := endTime.Hour()*60 + endTime.Minute()

	return domain.RestoreSchedule(idDB, roomDB, days, startMins, endMins, createdAt), nil
}
