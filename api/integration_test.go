package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simple-go/api/config"
	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/routes"

	"github.com/golang-jwt/jwt/v4"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	
	db.SetupTestConfig()
	db.SetupTestDatabase(t)

	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", routes.HealthHandler)
	mux.HandleFunc("/api/login", routes.LoginHandler)
	mux.HandleFunc("/api/register", routes.RegisterHandler)
	mux.HandleFunc("/api/users", middleware.AuthMiddleware(routes.UsersHandler))
	mux.HandleFunc("/api/profile", middleware.AuthMiddleware(routes.ProfileHandler))

	return httptest.NewServer(mux)
}

func getAuthToken(t *testing.T, server *httptest.Server) string {
	t.Helper()

	loginReq := map[string]string{
		"username": "admin",
		"password": "password",
	}
	body, _ := json.Marshal(loginReq)

	resp, err := http.Post(server.URL+"/api/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Login failed with status: %d", resp.StatusCode)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	return loginResp.Token
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Message != "Hello from Go service!" {
		t.Errorf("Expected message 'Hello from Go service!', got %s", response.Message)
	}
	if response.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", response.Status)
	}
}

func TestIntegration_LoginFlow(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	loginReq := map[string]string{
		"username": "admin",
		"password": "password",
	}
	body, _ := json.Marshal(loginReq)

	resp, err := http.Post(server.URL+"/api/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("Failed to decode login response: %v", err)
	}

	if loginResp.Token == "" {
		t.Error("Expected token in response")
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(loginResp.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		t.Errorf("Invalid token: %v", err)
	}

	if claims.Username != "admin" {
		t.Errorf("Expected username 'admin', got %s", claims.Username)
	}
}

func TestIntegration_RegisterAndLogin(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	registerReq := map[string]string{
		"username": "newuser",
		"password": "newpassword",
		"email":    "newuser@example.com",
	}
	body, _ := json.Marshal(registerReq)

	resp, err := http.Post(server.URL+"/api/register", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to register: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	loginReq := map[string]string{
		"username": "newuser",
		"password": "newpassword",
	}
	body, _ = json.Marshal(loginReq)

	resp, err = http.Post(server.URL+"/api/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestIntegration_ProtectedEndpoints(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	token := getAuthToken(t, server)

	testCases := []struct {
		name     string
		endpoint string
		method   string
	}{
		{"Users endpoint", "/api/users", "GET"},
		{"Profile endpoint", "/api/profile", "GET"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest(tc.method, server.URL+tc.endpoint, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to call %s: %v", tc.endpoint, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status %d for %s, got %d", http.StatusOK, tc.endpoint, resp.StatusCode)
			}
		})
	}
}

func TestIntegration_UnauthorizedAccess(t *testing.T) {
	server := setupTestServer(t)
	defer server.Close()

	testCases := []string{"/api/users", "/api/profile"}

	for _, endpoint := range testCases {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(server.URL + endpoint)
			if err != nil {
				t.Fatalf("Failed to call %s: %v", endpoint, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("Expected status %d for %s without auth, got %d", http.StatusUnauthorized, endpoint, resp.StatusCode)
			}
		})
	}
}