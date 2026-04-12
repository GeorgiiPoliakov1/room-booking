package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler_ServeHTTP(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}

	expected := "application/json"
	if contentType := rr.Header().Get("Content-Type"); contentType != expected {
		t.Errorf("expected Content-Type %s, got %s", expected, contentType)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("expected status 'healthy', got %s", response["status"])
	}
}
