package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"room-booking/internal/auth"
	"room-booking/internal/domain"
	"room-booking/internal/interface/dto/request"
	"room-booking/internal/interface/dto/response"
	"room-booking/internal/service"
)

type AuthHandler struct {
	jwtService *service.JWTService
}

func NewAuthHandler(jwtService *service.JWTService) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
	}
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {

	var req request.DummyLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	role := domain.Role(req.Role)

	if !role.IsValid() {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	var userID uuid.UUID

	if role == domain.RoleAdmin {
		userID = auth.AdminUserID
	} else {
		userID = auth.NormalUserID
	}

	token, err := h.jwtService.GenerateToken(userID, req.Role)
	if err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	resp := response.TokenResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
