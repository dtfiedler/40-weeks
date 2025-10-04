package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simple-go/config"

	"github.com/golang-jwt/jwt/v4"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
	config.AppConfig = &config.Config{
		JWTSecret: "test-secret",
	}

	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
}

func TestAuthMiddleware_InvalidHeaderFormat(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	config.AppConfig = &config.Config{
		JWTSecret: "test-secret",
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	config.AppConfig = &config.Config{
		JWTSecret: "test-secret",
	}

	claims := &Claims{
		UserID:   1,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if handlerCalled {
		t.Error("Expected handler not to be called")
	}
}