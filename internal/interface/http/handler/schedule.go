package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"room-booking/internal/domain"
	"room-booking/internal/service"
)

type CtxKey string

const (
	CtxRoomID CtxKey = "room_id"
)

type ScheduleCreator interface {
	Create(ctx context.Context, input service.CreateScheduleInput) (*domain.Schedule, error)
}

type ScheduleCreateHandler struct {
	svc ScheduleCreator
}

func NewScheduleHandler(svc ScheduleCreator) *ScheduleCreateHandler {
	return &ScheduleCreateHandler{svc: svc}
}

func (h *ScheduleCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	roomID, ok := r.Context().Value(CtxRoomID).(string)
	if !ok || roomID == "" {
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "roomId is required in path")
		return
	}
	var input service.CreateScheduleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "malformed JSON body")
		return
	}
	input.RoomID = roomID
	schedule, err := h.svc.Create(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, map[string]any{
		"schedule": map[string]any{
			"id":         schedule.ID(),
			"roomId":     schedule.RoomID(),
			"daysOfWeek": schedule.DaysOfWeek(),
			"startTime":  schedule.StartTimeUTC(),
			"endTime":    schedule.EndTimeUTC(),
		},
	})
}

func (h *ScheduleCreateHandler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRoomID),
		errors.Is(err, domain.ErrEmptyDaysOfWeek),
		errors.Is(err, domain.ErrInvalidDayOfWeek),
		errors.Is(err, domain.ErrDuplicateDayOfWeek),
		errors.Is(err, domain.ErrInvalidTimeFormat),
		errors.Is(err, domain.ErrEndTimeBeforeStart),
		errors.Is(err, domain.ErrSlotTooShort):
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", err.Error())

	case errors.Is(err, domain.ErrRoomNotFound):
		h.respondError(w, http.StatusNotFound, "NotFound", "room not found")

	case errors.Is(err, domain.ErrScheduleAlreadyExists):
		h.respondError(w, http.StatusConflict, "SCHEDULE_EXISTS", "schedule for this room already exists and cannot be changed")

	default:
		h.respondError(w, http.StatusInternalServerError, "InternalError", "internal server error")
	}
}

func (h *ScheduleCreateHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "service/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *ScheduleCreateHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "service/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
