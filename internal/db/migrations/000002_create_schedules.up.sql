CREATE TABLE IF NOT EXISTS schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL UNIQUE,
    days_of_week INTEGER[] NOT NULL CHECK (array_length(days_of_week, 1) > 0),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_time_order CHECK (end_time > start_time),
    CONSTRAINT chk_min_duration CHECK (EXTRACT(EPOCH FROM (end_time - start_time)) >= 1800)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_schedules_room_id ON schedules(room_id);