package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"simple-go/db"
)

func TestHealthHandler(t *testing.T) {
	db.SetupTestConfig()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	HealthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	expectedBody := `{"message":"Hello from Go service!","status":"healthy"}`
	actualBody := w.Body.String()
	if actualBody != expectedBody+"\n" {
		t.Errorf("Expected body %s, got %s", expectedBody, actualBody)
	}
}

func TestUsersHandler(t *testing.T) {
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()

	UsersHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestProfileHandler(t *testing.T) {
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	req := httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	ProfileHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}