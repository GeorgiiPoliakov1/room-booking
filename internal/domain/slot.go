package domain

import (
	"time"
)

type Slot struct {
	id     string
	roomID string
	start  time.Time
	end    time.Time
}

func NewSlot(id, roomID string, start, end time.Time) (*Slot, error) {
	start = start.UTC()
	end = end.UTC()

	if !end.After(start) {
		return nil, ErrEndTimeBeforeStart
	}
	if end.Sub(start) != 30*time.Minute {
		return nil, ErrInvalidSlotDuration
	}

	return &Slot{
		id:     id,
		roomID: roomID,
		start:  start,
		end:    end,
	}, nil
}

func (s *Slot) ID() string       { return s.id }
func (s *Slot) RoomID() string   { return s.roomID }
func (s *Slot) Start() time.Time { return s.start }
func (s *Slot) End() time.Time   { return s.end }

func (s *Slot) IsOccupiedBy(bookingStart, bookingEnd time.Time) bool {
	return bookingStart.Before(s.end) && bookingEnd.After(s.start)
}

func (s *Slot) IsWithinDate(date time.Time) bool {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)

	return (s.start.Equal(dayStart) || s.start.After(dayStart)) && s.start.Before(dayEnd)
}
