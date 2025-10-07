package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"simple-go/api/db"
)

func TestLoginHandler_Success(t *testing.T) {
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	loginReq := LoginRequest{
		Username: "admin",
		Password: "password",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response, got empty string")
	}
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	loginReq := LoginRequest{
		Username: "admin",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLoginHandler_NonexistentUser(t *testing.T) {
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	loginReq := LoginRequest{
		Username: "nonexistent",
		Password: "password",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLoginHandler_InvalidMethod(t *testing.T) {
	db.SetupTestConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	db.SetupTestConfig()

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	LoginHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegisterHandler_Success(t *testing.T) {
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	registerReq := RegisterRequest{
		Username: "testuser",
		Password: "testpassword",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(registerReq)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["message"] != "User created successfully" {
		t.Errorf("Expected success message, got %s", response["message"])
	}
}

func TestRegisterHandler_DuplicateUser(t *testing.T) {
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	registerReq := RegisterRequest{
		Username: "admin",
		Password: "testpassword",
		Email:    "admin2@example.com",
	}
	body, _ := json.Marshal(registerReq)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestRegisterHandler_MissingFields(t *testing.T) {
	db.SetupTestConfig()

	registerReq := RegisterRequest{
		Username: "testuser",
		Password: "",
		Email:    "test@example.com",
	}
	body, _ := json.Marshal(registerReq)

	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestRegisterHandler_InvalidMethod(t *testing.T) {
	db.SetupTestConfig()

	req := httptest.NewRequest(http.MethodGet, "/api/register", nil)
	w := httptest.NewRecorder()

	RegisterHandler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}