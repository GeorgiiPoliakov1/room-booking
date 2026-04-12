package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"room-booking/internal/domain"
	"room-booking/internal/interface/dto/request"
	"room-booking/internal/interface/dto/response"
	"room-booking/internal/service"
)

type BookingCreator interface {
	Create(ctx context.Context, input service.BookingCreateInput) (*domain.Booking, error)
}

type BookingCreateHandler struct {
	service BookingCreator
}

func NewBookingCreateHandler(svc BookingCreator) *BookingCreateHandler {
	return &BookingCreateHandler{service: svc}
}

func (h *BookingCreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	userID, ok := r.Context().Value(CtxUserID).(string)
	if !ok || userID == "" {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized", "authentication required")
		return
	}

	var req request.BookingCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "malformed JSON body")
		return
	}

	needConfLink := req.CreateConferenceLink != nil && *req.CreateConferenceLink

	input := service.BookingCreateInput{
		SlotID:               req.SlotID,
		UserID:               userID,
		CreateConferenceLink: needConfLink,
	}

	booking, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, response.BookingCreateResponse{
		Booking: mapBookingToResponse(booking),
	})
}

func (h *BookingCreateHandler) handleDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidSlotID):
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "invalid slot ID format")
	case errors.Is(err, domain.ErrInvalidUserID):
		h.respondError(w, http.StatusBadRequest, "InvalidRequest", "invalid user ID format")
	case errors.Is(err, domain.ErrSlotInPast):
		h.respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "slot start time is in the past")
	case errors.Is(err, domain.ErrSlotNotFound):
		h.respondError(w, http.StatusNotFound, "NotFound", "slot not found")
	case errors.Is(err, domain.ErrSlotAlreadyBooked):
		h.respondError(w, http.StatusConflict, "SLOT_ALREADY_BOOKED", "slot is already booked")
	default:
		h.respondError(w, http.StatusInternalServerError, "InternalError", "internal server error")
	}
}

func mapBookingToResponse(b *domain.Booking) response.Booking {
	resp := response.Booking{
		ID:     b.ID(),
		SlotID: b.SlotID(),
		UserID: b.UserID(),
		Status: string(b.Status()),
	}

	if link := b.ConferenceLink(); link != nil {
		resp.ConferenceLink = link
	}
	createdAt := b.CreatedAt()
	if !createdAt.IsZero() {
		resp.CreatedAt = &createdAt
	}

	return resp
}

func (h *BookingCreateHandler) respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *BookingCreateHandler) respondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
