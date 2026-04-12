package response

import "time"

type ScheduleResponse struct {
	DaysOfWeek []int     `json:"days_of_week"`
	StartTime  string    `json:"start_time"`
	EndTime    string    `json:"end_time"`
	CreatedAt  time.Time `json:"created_at"`
}

type ScheduleCreateResponse struct {
	Schedule ScheduleResponse `json:"schedule"`
}
