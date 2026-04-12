package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewSchedule(t *testing.T) {
	tests := []struct {
		name        string
		params      ScheduleCreateParams
		expectErr   error
		expectStart string
		expectEnd   string
		expectDur   int
	}{
		{
			name: "success: single day, standard hours",
			params: ScheduleCreateParams{
				ID:         "550e8400-e29b-41d4-a716-446655440000",
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{1},
				StartTime:  "09:00",
				EndTime:    "10:00",
			},
			expectStart: "09:00",
			expectEnd:   "10:00",
			expectDur:   60,
		},
		{
			name: "success: multiple days, cross noon",
			params: ScheduleCreateParams{
				ID:         "",
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{1, 3, 5, 7},
				StartTime:  "13:30",
				EndTime:    "15:00",
			},
			expectStart: "13:30",
			expectEnd:   "15:00",
			expectDur:   90,
		},
		{
			name: "success: minimum allowed duration (30 mins)",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{2},
				StartTime:  "08:00",
				EndTime:    "08:30",
			},
			expectStart: "08:00",
			expectEnd:   "08:30",
			expectDur:   30,
		},
		{
			name: "fail: empty roomID",
			params: ScheduleCreateParams{
				RoomID:     "",
				DaysOfWeek: []int{1},
				StartTime:  "09:00",
				EndTime:    "10:00",
			},
			expectErr: ErrInvalidRoomID,
		},
		{
			name: "fail: empty daysOfWeek",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{},
				StartTime:  "09:00",
				EndTime:    "10:00",
			},
			expectErr: ErrEmptyDaysOfWeek,
		},
		{
			name: "fail: day out of range (0)",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{0},
				StartTime:  "09:00",
				EndTime:    "10:00",
			},
			expectErr: ErrInvalidDayOfWeek,
		},
		{
			name: "fail: day out of range (8)",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{8},
				StartTime:  "09:00",
				EndTime:    "10:00",
			},
			expectErr: ErrInvalidDayOfWeek,
		},
		{
			name: "fail: end time before start time",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{1},
				StartTime:  "10:00",
				EndTime:    "09:00",
			},
			expectErr: ErrEndTimeBeforeStart,
		},
		{
			name: "fail: end time equal to start time",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{1},
				StartTime:  "10:00",
				EndTime:    "10:00",
			},
			expectErr: ErrEndTimeBeforeStart,
		},
		{
			name: "fail: duration less than 30 minutes",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{1},
				StartTime:  "10:00",
				EndTime:    "10:29",
			},
			expectErr: ErrSlotTooShort,
		},
		{
			name: "fail: invalid time format",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{1},
				StartTime:  "25:00",
				EndTime:    "26:00",
			},
			expectErr: ErrInvalidTimeFormat,
		},
		{
			name: "fail: dublicates in daysOfWeek",
			params: ScheduleCreateParams{
				RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
				DaysOfWeek: []int{1, 2, 2},
				StartTime:  "21:00",
				EndTime:    "23:00",
			},
			expectErr: ErrDuplicateDayOfWeek,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := NewSchedule(tt.params)

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

			if schedule.RoomID() != tt.params.RoomID {
				t.Errorf("RoomID: expected %s, got %s", tt.params.RoomID, schedule.RoomID())
			}
			if len(schedule.DaysOfWeek()) != len(tt.params.DaysOfWeek) {
				t.Errorf("DaysOfWeek length: expected %d, got %d", len(tt.params.DaysOfWeek), len(schedule.DaysOfWeek()))
			}
			if schedule.StartTimeUTC() != tt.expectStart {
				t.Errorf("StartTimeUTC: expected %s, got %s", tt.expectStart, schedule.StartTimeUTC())
			}
			if schedule.EndTimeUTC() != tt.expectEnd {
				t.Errorf("EndTimeUTC: expected %s, got %s", tt.expectEnd, schedule.EndTimeUTC())
			}
			if schedule.DurationMinutes() != tt.expectDur {
				t.Errorf("DurationMinutes: expected %d, got %d", tt.expectDur, schedule.DurationMinutes())
			}
			if schedule.CreatedAt().IsZero() {
				t.Error("CreatedAt should be set")
			}
			if schedule.CreatedAt().Location() != time.UTC {
				t.Error("CreatedAt must be in UTC")
			}
		})
	}
}

func TestScheduleImmutability(t *testing.T) {
	params := ScheduleCreateParams{
		RoomID:     "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
		DaysOfWeek: []int{1, 3, 5},
		StartTime:  "09:00",
		EndTime:    "10:30",
	}
	schedule, err := NewSchedule(params)
	if err != nil {
		t.Fatalf("failed to create schedule: %v", err)
	}

	days := schedule.DaysOfWeek()
	days[0] = 99
	days = append(days, 7)

	if schedule.DaysOfWeek()[0] == 99 {
		t.Error("internal daysOfWeek slice was mutated via getter")
	}
	if len(schedule.DaysOfWeek()) != 3 {
		t.Errorf("internal slice length changed after external append: got %d", len(schedule.DaysOfWeek()))
	}

}
