package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"simple-go/api/config"
	"simple-go/api/db"
	"simple-go/api/middleware"
)

// UploadCoverPhotoHandler handles uploading a cover photo for a pregnancy
func UploadCoverPhotoHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Parse multipart form data
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, fileHeader, err := r.FormFile("cover_photo")
	if err != nil {
		http.Error(w, "No file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		http.Error(w, "Invalid file type. Only JPG, PNG, and WebP are allowed", http.StatusBadRequest)
		return
	}

	// Get pregnancy for the user
	var pregnancyID int
	err = db.GetDB().QueryRow(`
		SELECT id FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE
		ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&pregnancyID)
	if err != nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Create covers directory if it doesn't exist
	coversDir := filepath.Join(config.AppConfig.ImagesDirectory, "covers")
	if err := os.MkdirAll(coversDir, 0755); err != nil {
		http.Error(w, "Failed to create covers directory", http.StatusInternalServerError)
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("pregnancy_%d_cover_%d%s", pregnancyID, time.Now().Unix(), ext)
	fullPath := filepath.Join(coversDir, filename)

	// Create the file
	dst, err := os.Create(fullPath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file
	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Update pregnancy with new cover photo filename
	_, err = db.GetDB().Exec(`
		UPDATE pregnancies 
		SET cover_photo_filename = ?, updated_at = ?
		WHERE id = ?`,
		filename, time.Now(), pregnancyID)
	if err != nil {
		// Clean up the file if database update fails
		os.Remove(fullPath)
		http.Error(w, "Failed to update pregnancy", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"success": true, "filename": "%s"}`, filename)))
}

// DeleteCoverPhotoHandler removes the cover photo from a pregnancy
func DeleteCoverPhotoHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get pregnancy and current cover photo
	var pregnancyID int
	var currentFilename *string
	err := db.GetDB().QueryRow(`
		SELECT id, cover_photo_filename FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE
		ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&pregnancyID, &currentFilename)
	if err != nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Remove file if it exists
	if currentFilename != nil && *currentFilename != "" {
		coversDir := filepath.Join(config.AppConfig.ImagesDirectory, "covers")
		fullPath := filepath.Join(coversDir, *currentFilename)
		os.Remove(fullPath) // Ignore errors if file doesn't exist
	}

	// Update pregnancy to remove cover photo
	_, err = db.GetDB().Exec(`
		UPDATE pregnancies 
		SET cover_photo_filename = NULL, updated_at = ?
		WHERE id = ?`,
		time.Now(), pregnancyID)
	if err != nil {
		http.Error(w, "Failed to update pregnancy", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}