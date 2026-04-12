package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"room-booking/internal/domain"
	"room-booking/internal/interface/dto/response"
)

type SlotLister interface {
	GetAvailableSlots(ctx context.Context, roomID, dateStr string) ([]domain.Slot, error)
}

type SlotListHandler struct {
	service SlotLister
}

func NewSlotListHandler(svc SlotLister) *SlotListHandler {
	return &SlotListHandler{service: svc}
}

func (h *SlotListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	roomID, ok := r.Context().Value(CtxRoomID).(string)
	if !ok || roomID == "" {
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "roomId is required in path")
		return
	}

	dateStr := r.URL.Query().Get("date")
	if dateStr == "" {
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "query parameter 'date' is required")
		return
	}
	if _, err := time.Parse("2006-01-02", dateStr); err != nil {
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "date must be in YYYY-MM-DD format")
		return
	}

	slots, err := h.service.GetAvailableSlots(r.Context(), roomID, dateStr)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	slotsResp := make([]response.Slot, len(slots))
	for i, s := range slots {
		slotsResp[i] = response.Slot{
			ID:     s.ID(),
			RoomID: s.RoomID(),
			Start:  s.Start().UTC(),
			End:    s.End().UTC(),
		}
	}

	h.respondJSON(w, http.StatusOK, response.SlotListResponse{Slots: slotsResp})
}

func (h *SlotListHandler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidRoomID):
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "invalid room ID format")
	case errors.Is(err, domain.ErrRoomNotFound):
		h.respondError(w, http.StatusNotFound, "NotFound", "room not found")
	case err != nil && err.Error() != "":
		if err.Error()[:18] == "invalid date format" {
			h.respondError(w, http.StatusBadRequest, "InvalidRequest", "date must be in YYYY-MM-DD format")
			return
		}
		fallthrough
	default:
		h.respondError(w, http.StatusInternalServerError, "InternalError", "internal server error")
	}
}

func (h *SlotListHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *SlotListHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
