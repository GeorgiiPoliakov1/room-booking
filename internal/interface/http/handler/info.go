package handler

import (
	"encoding/json"
	"net/http"
)

type InfoHandler struct{}

func NewInfoHandler() *InfoHandler {
	return &InfoHandler{}
}

func (h *InfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"status":  "ok",
		"message": "Service is running",
	}
	_ = json.NewEncoder(w).Encode(response)
}
