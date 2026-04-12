package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"room-booking/internal/interface/dto/response"
	"room-booking/internal/service"

	"github.com/golang-jwt/jwt"
)

func TestDummyLogin_Admin(t *testing.T) {

	jwtService := service.NewJWTService("test-secret")
	handler := NewAuthHandler(jwtService)

	body := []byte(`{"role":"admin"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/dummyLogin",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler.DummyLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rec.Code)
	}
}

func TestDummyLogin_User(t *testing.T) {

	jwtService := service.NewJWTService("test-secret")
	handler := NewAuthHandler(jwtService)

	body := []byte(`{"role":"user"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/dummyLogin",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler.DummyLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rec.Code)
	}

	var resp response.TokenResponse

	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected token in response")
	}
}

func TestDummyLogin_InvalidRole(t *testing.T) {

	jwtService := service.NewJWTService("test-secret")
	handler := NewAuthHandler(jwtService)

	body := []byte(`{"role":"manager"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/dummyLogin",
		bytes.NewBuffer(body),
	)

	rec := httptest.NewRecorder()

	handler.DummyLogin(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 got %d", rec.Code)
	}
}

func TestDummyLogin_JSONResponse(t *testing.T) {

	jwtService := service.NewJWTService("test-secret")
	handler := NewAuthHandler(jwtService)

	body := []byte(`{"role":"admin"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/dummyLogin",
		bytes.NewBuffer(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	handler.DummyLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 got %d", rec.Code)
	}

	var resp map[string]string

	err := json.NewDecoder(rec.Body).Decode(&resp)
	if err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}

	token, ok := resp["token"]
	if !ok {
		t.Fatal("token field missing in response")
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}
}

func TestDummyLogin_JWTClaims(t *testing.T) {

	secret := "test-secret"

	jwtService := service.NewJWTService(secret)
	handler := NewAuthHandler(jwtService)

	body := []byte(`{"role":"admin"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/dummyLogin",
		bytes.NewBuffer(body),
	)

	rec := httptest.NewRecorder()

	handler.DummyLogin(rec, req)

	var resp response.TokenResponse
	json.NewDecoder(rec.Body).Decode(&resp)

	token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}

	claims := token.Claims.(jwt.MapClaims)

	if claims["role"] != "admin" {
		t.Fatalf("expected role admin got %v", claims["role"])
	}
}
