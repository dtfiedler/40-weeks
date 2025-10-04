package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupTestFiles(t *testing.T) {
	t.Helper()
	
	testDir := "test_public"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFiles := map[string]string{
		"login.html":     "<html><body><h1>Login Page</h1></body></html>",
		"register.html":  "<html><body><h1>Register Page</h1></body></html>",
		"dashboard.html": "<html><body><h1>Dashboard</h1></body></html>",
	}

	for filename, content := range testFiles {
		filepath := filepath.Join(testDir, filename)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	originalDir, _ := os.Getwd()
	os.Chdir("../")
	
	if err := os.RemoveAll("public"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to remove existing public directory: %v", err)
	}
	
	if err := os.Rename(testDir, "public"); err != nil {
		t.Fatalf("Failed to rename test directory: %v", err)
	}

	t.Cleanup(func() {
		os.Chdir(originalDir)
		os.RemoveAll("../public")
		
		originalPublicFiles := map[string]string{
			"login.html":     "<!-- Original login content -->",
			"register.html":  "<!-- Original register content -->", 
			"dashboard.html": "<!-- Original dashboard content -->",
		}
		
		os.MkdirAll("../public", 0755)
		for filename, content := range originalPublicFiles {
			filepath := filepath.Join("../public", filename)
			os.WriteFile(filepath, []byte(content), 0644)
		}
	})
}

func TestLoginPageHandler(t *testing.T) {
	setupTestFiles(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	LoginPageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html; charset=utf-8, got %s", contentType)
	}

	body := w.Body.String()
	if body != "<html><body><h1>Login Page</h1></body></html>" {
		t.Errorf("Unexpected body content: %s", body)
	}
}

func TestRegisterPageHandler(t *testing.T) {
	setupTestFiles(t)

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	w := httptest.NewRecorder()

	RegisterPageHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html; charset=utf-8, got %s", contentType)
	}

	body := w.Body.String()
	if body != "<html><body><h1>Register Page</h1></body></html>" {
		t.Errorf("Unexpected body content: %s", body)
	}
}

func TestDashboardHandler(t *testing.T) {
	setupTestFiles(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()

	DashboardHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type text/html; charset=utf-8, got %s", contentType)
	}

	body := w.Body.String()
	if body != "<html><body><h1>Dashboard</h1></body></html>" {
		t.Errorf("Unexpected body content: %s", body)
	}
}

func TestStaticHandlers_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	w := httptest.NewRecorder()

	LoginPageHandler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d when file not found, got %d", http.StatusNotFound, w.Code)
	}
}