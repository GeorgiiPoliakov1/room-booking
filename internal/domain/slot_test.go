package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewSlot(t *testing.T) {
	utc := time.UTC
	mkTime := func(h, m int) time.Time {
		return time.Date(2024, 6, 10, h, m, 0, 0, utc)
	}

	tests := []struct {
		name      string
		id        string
		roomID    string
		start     time.Time
		end       time.Time
		expectErr error
		checkUTC  bool
	}{
		{
			name: "success: valid 30min slot",
			id:   "slot-1", roomID: "room-1",
			start: mkTime(9, 0), end: mkTime(9, 30),
		},
		{
			name: "success: converts non-UTC to UTC",
			id:   "slot-2", roomID: "room-2",
			start:    time.Date(2024, 6, 10, 12, 0, 0, 0, time.FixedZone("UTC+3", 3*60*60)),
			end:      time.Date(2024, 6, 10, 12, 30, 0, 0, time.FixedZone("UTC+3", 3*60*60)),
			checkUTC: true,
		},
		{
			name: "fail: end time equals start time",
			id:   "slot-3", roomID: "room-1",
			start: mkTime(9, 0), end: mkTime(9, 0),
			expectErr: ErrEndTimeBeforeStart,
		},
		{
			name: "fail: end time before start time",
			id:   "slot-4", roomID: "room-1",
			start: mkTime(9, 30), end: mkTime(9, 0),
			expectErr: ErrEndTimeBeforeStart,
		},
		{
			name: "fail: duration 29 minutes",
			id:   "slot-5", roomID: "room-1",
			start: mkTime(9, 0), end: mkTime(9, 29),
			expectErr: ErrInvalidSlotDuration,
		},
		{
			name: "fail: duration 60 minutes",
			id:   "slot-6", roomID: "room-1",
			start: mkTime(9, 0), end: mkTime(10, 0),
			expectErr: ErrInvalidSlotDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot, err := NewSlot(tt.id, tt.roomID, tt.start, tt.end)

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectErr)
				}
				if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if slot.ID() != tt.id {
				t.Errorf("ID mismatch: expected %s, got %s", tt.id, slot.ID())
			}
			if slot.RoomID() != tt.roomID {
				t.Errorf("RoomID mismatch: expected %s, got %s", tt.roomID, slot.RoomID())
			}

			if tt.checkUTC {
				if slot.Start().Location() != time.UTC {
					t.Error("Start time should be normalized to UTC")
				}
				if slot.End().Location() != time.UTC {
					t.Error("End time should be normalized to UTC")
				}
			}

			if !slot.Start().Equal(tt.start.UTC()) {
				t.Errorf("Start mismatch: expected %v, got %v", tt.start.UTC(), slot.Start())
			}
			if !slot.End().Equal(tt.end.UTC()) {
				t.Errorf("End mismatch: expected %v, got %v", tt.end.UTC(), slot.End())
			}
		})
	}
}

func TestSlot_IsOccupiedBy(t *testing.T) {
	utc := time.UTC
	mkTime := func(h, m int) time.Time { return time.Date(2024, 6, 10, h, m, 0, 0, utc) }

	slot, err := NewSlot("s1", "r1", mkTime(9, 0), mkTime(9, 30))
	if err != nil {
		t.Fatalf("failed to create test slot: %v", err)
	}

	tests := []struct {
		name        string
		bookStart   time.Time
		bookEnd     time.Time
		expectOccup bool
	}{
		{"no overlap: booking completely before", mkTime(8, 0), mkTime(8, 30), false},
		{"no overlap: booking completely after", mkTime(10, 0), mkTime(10, 30), false},
		{"no overlap: booking ends exactly at slot start", mkTime(8, 30), mkTime(9, 0), false},
		{"no overlap: booking starts exactly at slot end", mkTime(9, 30), mkTime(10, 0), false},

		{"overlap: booking covers entire slot", mkTime(8, 0), mkTime(10, 0), true},
		{"overlap: booking starts inside slot", mkTime(9, 15), mkTime(9, 45), true},
		{"overlap: booking ends inside slot", mkTime(8, 45), mkTime(9, 15), true},
		{"overlap: exact match", mkTime(9, 0), mkTime(9, 30), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slot.IsOccupiedBy(tt.bookStart, tt.bookEnd)
			if got != tt.expectOccup {
				t.Errorf("IsOccupiedBy(%v, %v) = %v; want %v", tt.bookStart, tt.bookEnd, got, tt.expectOccup)
			}
		})
	}
}

func TestSlot_IsWithinDate(t *testing.T) {
	utc := time.UTC
	targetDate := time.Date(2024, 6, 10, 0, 0, 0, 0, utc)

	tests := []struct {
		name      string
		slotStart time.Time
		queryDate time.Time
		expectIn  bool
	}{
		{"match: slot starts at 09:00 on target date", time.Date(2024, 6, 10, 9, 0, 0, 0, utc), targetDate, true},
		{"match: slot starts at 23:30 on target date", time.Date(2024, 6, 10, 23, 30, 0, 0, utc), targetDate, true},
		{"no match: slot starts previous day 23:30", time.Date(2024, 6, 9, 23, 30, 0, 0, utc), targetDate, false},
		{"no match: slot starts next day 00:00", time.Date(2024, 6, 11, 0, 0, 0, 0, utc), targetDate, false},
		{"match: query date has non-zero time (should be ignored)", time.Date(2024, 6, 10, 14, 0, 0, 0, utc), time.Date(2024, 6, 10, 15, 30, 0, 0, utc), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slot, _ := NewSlot("s1", "r1", tt.slotStart, tt.slotStart.Add(30*time.Minute))
			got := slot.IsWithinDate(tt.queryDate)
			if got != tt.expectIn {
				t.Errorf("IsWithinDate(%v) = %v; want %v", tt.queryDate, got, tt.expectIn)
			}
		})
	}
}
