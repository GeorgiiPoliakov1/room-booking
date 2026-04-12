CREATE TABLE IF NOT EXISTS slots (
    id UUID PRIMARY KEY,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_slot_duration CHECK (EXTRACT(EPOCH FROM (end_time - start_time)) = 1800),
    CONSTRAINT chk_slot_time_order CHECK (end_time > start_time),
    CONSTRAINT uq_slot_room_start UNIQUE (room_id, start_time)
);

CREATE INDEX idx_slots_room_date ON slots (room_id, start_time DESC);