package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInfoHandler_ServeHTTP(t *testing.T) {
	handler := NewInfoHandler()

	req := httptest.NewRequest(http.MethodGet, "/_info", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status OK, got %v", status)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status 'ok', got %s", response["status"])
	}

	if response["message"] != "Service is running" {
		t.Errorf("expected message 'Service is running', got %s", response["message"])
	}
}
