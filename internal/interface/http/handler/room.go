package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"

	"room-booking/internal/domain"
	"room-booking/internal/interface/dto/request"
	"room-booking/internal/interface/dto/response"
	"room-booking/internal/service"
)

type ctxKey string

const (
	CtxUserID   ctxKey = "user_id"
	CtxUserRole ctxKey = "user_role"
)

type RoomHandler struct {
	service  service.RoomService
	validate *validator.Validate
	logger   Logger
}

type Logger interface {
	Error(msg string, args ...any)
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
}

func NewRoomHandler(svc service.RoomService, logger Logger) *RoomHandler {
	validate := validator.New(validator.WithRequiredStructEnabled())

	_ = validate.RegisterValidation("room_name", func(fl validator.FieldLevel) bool {
		name := fl.Field().String()
		return len(name) > 0 && len(name) <= domain.MaxRoomNameLength
	})

	return &RoomHandler{
		service:  svc,
		validate: validate,
		logger:   logger,
	}
}

func (h *RoomHandler) ListRooms(w http.ResponseWriter, r *http.Request) {
	const op = "handler.RoomHandler.ListRooms"

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req request.RoomListRequest
	req.Page = 1
	req.Limit = 20

	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &req.Page)
	}
	if l := r.URL.Query().Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &req.Limit)
	}

	if err := h.validate.Struct(req); err != nil {
		h.respondValidationError(w, err)
		return
	}

	limit := req.Limit
	offset := (req.Page - 1) * limit

	rooms, err := h.service.ListRooms(ctx, limit, offset)
	if err != nil {
		h.logger.Error(op, "err", err, "page", req.Page, "limit", req.Limit)
		h.respondJSON(w, http.StatusInternalServerError, response.ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	roomResponses := make([]response.RoomResponse, 0, len(rooms))
	for _, room := range rooms {
		if room == nil {
			continue
		}
		roomResponses = append(roomResponses, response.RoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			Capacity:    room.Capacity,
			CreatedAt:   room.CreatedAt,
			UpdatedAt:   room.UpdatedAt,
		})
	}
	h.respondJSON(w, http.StatusOK, response.RoomsListResponse{
		Rooms: roomResponses,
		Total: len(roomResponses),
		Page:  req.Page,
		Limit: req.Limit,
	})
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	const op = "handler.RoomHandler.CreateRoom"
	role, ok := r.Context().Value(CtxUserRole).(string)
	if !ok || role != "admin" {
		h.respondJSON(w, http.StatusForbidden, response.ErrorResponse{
			Error: "access denied: admin role required",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	defer r.Body.Close()

	var req request.RoomCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondJSON(w, http.StatusBadRequest, response.ErrorResponse{
			Error: "malformed json",
		})
		return
	}
	if err := h.validate.Struct(req); err != nil {
		h.respondValidationError(w, err)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	room, err := h.service.CreateRoom(ctx, req.Name, req.Description, req.Capacity)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRoomNameExists):
			h.respondJSON(w, http.StatusBadRequest, response.ErrorResponse{
				Error: "room name already exists",
			})
		case errors.Is(err, domain.ErrRoomNameRequired),
			errors.Is(err, domain.ErrRoomNameTooLong),
			errors.Is(err, domain.ErrDescriptionTooLong),
			errors.Is(err, domain.ErrCapacityOutOfRange):
			h.respondJSON(w, http.StatusBadRequest, response.ErrorResponse{
				Error: err.Error(),
			})
		default:
			h.logger.Error(op, "err", err, "name", req.Name)
			h.respondJSON(w, http.StatusInternalServerError, response.ErrorResponse{
				Error: "failed to create room",
			})
		}
		return
	}

	h.respondJSON(w, http.StatusCreated, response.RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		Description: room.Description,
		Capacity:    room.Capacity,
		CreatedAt:   room.CreatedAt,
		UpdatedAt:   room.UpdatedAt,
	})
}

func (h *RoomHandler) respondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func (h *RoomHandler) respondValidationError(w http.ResponseWriter, err error) {
	resp := response.ErrorResponse{Error: "validation failed"}

	if ve, ok := err.(validator.ValidationErrors); ok {
		resp.Details = make([]response.ValidationError, 0, len(ve))
		for _, e := range ve {
			resp.Details = append(resp.Details, response.ValidationError{
				Field:   e.Field(),
				Message: h.validationMessage(e),
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *RoomHandler) validationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required"
	case "min":
		return fmt.Sprintf("minimum value is %s", e.Param())
	case "max":
		return fmt.Sprintf("maximum value is %s", e.Param())
	case "datetime":
		return fmt.Sprintf("invalid time format, expected %s", e.Param())
	case "gtfield":
		return fmt.Sprintf("must be greater than field %s", e.Param())
	case "room_name":
		return "invalid room name format"
	case "omitempty":
		return ""
	default:
		return "invalid value"
	}
}
