package domain

import (
	"fmt"
	"regexp"
	"time"
)

var timeHHMMRegex = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):[0-5][0-9]$`)

type Schedule struct {
	id         string
	roomID     string
	daysOfWeek []int
	startTime  int
	endTime    int
	createdAt  time.Time
}

type ScheduleCreateParams struct {
	ID         string
	RoomID     string
	DaysOfWeek []int
	StartTime  string
	EndTime    string
}

func NewSchedule(params ScheduleCreateParams) (*Schedule, error) {
	if params.ID != "" && !isValidUUID(params.ID) {
		return nil, ErrInvalidScheduleID
	}

	if params.RoomID == "" || !isValidUUID(params.RoomID) {
		return nil, ErrInvalidRoomID
	}

	if len(params.DaysOfWeek) == 0 {
		return nil, ErrEmptyDaysOfWeek
	}

	for _, d := range params.DaysOfWeek {
		if d < 1 || d > 7 {
			return nil, fmt.Errorf("%w: got %d", ErrInvalidDayOfWeek, d)
		}
	}

	seen := make(map[int]struct{}, len(params.DaysOfWeek))
	for _, d := range params.DaysOfWeek {
		if _, exists := seen[d]; exists {
			return nil, fmt.Errorf("%w: %d", ErrDuplicateDayOfWeek, d)
		}
		seen[d] = struct{}{}
	}

	startMins, err := parseTimeToMinutes(params.StartTime)
	if err != nil {
		return nil, fmt.Errorf("start time: %w", err)
	}
	endMins, err := parseTimeToMinutes(params.EndTime)
	if err != nil {
		return nil, fmt.Errorf("end time: %w", err)
	}

	if endMins <= startMins {
		return nil, ErrEndTimeBeforeStart
	}

	if endMins-startMins < 30 {
		return nil, ErrSlotTooShort
	}

	daysCopy := make([]int, len(params.DaysOfWeek))
	copy(daysCopy, params.DaysOfWeek)

	return &Schedule{
		id:         params.ID,
		roomID:     params.RoomID,
		daysOfWeek: daysCopy,
		startTime:  startMins,
		endTime:    endMins,
		createdAt:  time.Now().UTC(),
	}, nil
}

func (s *Schedule) ID() string           { return s.id }
func (s *Schedule) RoomID() string       { return s.roomID }
func (s *Schedule) DaysOfWeek() []int    { return append([]int(nil), s.daysOfWeek...) }
func (s *Schedule) StartTimeUTC() string { return minutesToTimeStr(s.startTime) }
func (s *Schedule) EndTimeUTC() string   { return minutesToTimeStr(s.endTime) }
func (s *Schedule) CreatedAt() time.Time { return s.createdAt }
func (s *Schedule) DurationMinutes() int { return s.endTime - s.startTime }

func RestoreSchedule(id, roomID string, daysOfWeek []int, startMin, endMin int, createdAt time.Time) *Schedule {
	daysCopy := make([]int, len(daysOfWeek))
	copy(daysCopy, daysOfWeek)

	return &Schedule{
		id:         id,
		roomID:     roomID,
		daysOfWeek: daysCopy,
		startTime:  startMin,
		endTime:    endMin,
		createdAt:  createdAt,
	}
}

func parseTimeToMinutes(hhmm string) (int, error) {
	if !timeHHMMRegex.MatchString(hhmm) {
		return 0, ErrInvalidTimeFormat
	}
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidTimeFormat, err)
	}
	return t.Hour()*60 + t.Minute(), nil
}

func minutesToTimeStr(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	return fmt.Sprintf("%02d:%02d", h, m)
}

func (s *Schedule) IsAvailableOn(dayOfWeek int) bool {
	for _, d := range s.daysOfWeek {
		if d == dayOfWeek {
			return true
		}
	}
	return false
}
func (s *Schedule) IsTimeInRange(timeMinutes int) bool {
	return timeMinutes >= s.startTime && timeMinutes < s.endTime
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}

	for i, r := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
			continue
		}
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
