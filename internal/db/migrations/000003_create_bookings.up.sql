CREATE TABLE IF NOT EXISTS bookings (
    id UUID PRIMARY KEY,
    slot_id UUID NOT NULL,
    user_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    conference_link TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_booking_status CHECK (status IN ('active', 'cancelled'))
);

CREATE UNIQUE INDEX idx_bookings_active_slot ON bookings (slot_id) WHERE status = 'active';

CREATE INDEX idx_bookings_user_id ON bookings (user_id);
CREATE INDEX idx_bookings_created_at ON bookings (created_at DESC);