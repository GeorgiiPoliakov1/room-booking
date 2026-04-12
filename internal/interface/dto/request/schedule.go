package request

type ScheduleCreateRequest struct {
	RoomID     string `json:"roomId" validate:"required,uuid_custom"`
	DaysOfWeek []int  `json:"daysOfWeek" validate:"required,dive,gte=1,lte=7"`
	StartTime  string `json:"startTime" validate:"required,time_hhmm"`
	EndTime    string `json:"endTime" validate:"required,time_hhmm"`
}
