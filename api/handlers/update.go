package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"simple-go/api/config"
	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
	"simple-go/api/services/email"
)

// CreateUpdateRequest represents the request to create a new pregnancy update
type CreateUpdateRequest struct {
	Title           string  `json:"title"`
	Content         *string `json:"content"`
	UpdateType      string  `json:"update_type"`
	AppointmentType *string `json:"appointment_type"`
	IsShared        bool    `json:"is_shared"`
	Date            *string `json:"date"` // ISO string for update date, defaults to current UTC time
}

// CreateUpdateHandler handles creating a new pregnancy update
func CreateUpdateHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Parse multipart form data for file uploads
	err := r.ParseMultipartForm(100 << 20) // 100 MB max for videos
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Get pregnancy for the user
	var pregnancyID int
	var conceptionDate *time.Time
	err = db.GetDB().QueryRow(`
		SELECT id, conception_date FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE
		ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&pregnancyID, &conceptionDate)
	if err != nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Parse JSON data from form field
	var req CreateUpdateRequest
	jsonData := r.FormValue("data")
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}
	}

	// Validate required fields
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Parse update date or use current UTC time
	var updateDate time.Time
	if req.Date != nil && *req.Date != "" {
		parsedDate, err := time.Parse(time.RFC3339, *req.Date)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		updateDate = parsedDate.UTC()
	} else {
		updateDate = time.Now().UTC()
	}

	// Calculate week number based on conception date
	var weekNumber *int
	if conceptionDate != nil {
		// Calculate gestational weeks from LMP (conception + 14 days)
		daysSinceConception := int(updateDate.Sub(*conceptionDate).Hours() / 24)
		// Add 14 days to account for LMP offset (gestational age calculation)
		calculatedWeek := (daysSinceConception+14)/7 + 1
		if calculatedWeek > 0 {
			weekNumber = &calculatedWeek
		}
	}

	// Insert the update
	result, err := db.GetDB().Exec(`
		INSERT INTO pregnancy_updates (pregnancy_id, week_number, title, content, update_type, appointment_type, is_shared, shared_at, update_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		pregnancyID, weekNumber, req.Title, req.Content, req.UpdateType, req.AppointmentType, req.IsShared,
		func() *time.Time {
			if req.IsShared {
				now := time.Now()
				return &now
			}
			return nil
		}(), &updateDate)

	if err != nil {
		http.Error(w, "Failed to create update", http.StatusInternalServerError)
		return
	}

	updateID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to get update ID", http.StatusInternalServerError)
		return
	}

	// Handle photo uploads
	files := r.MultipartForm.File["photos"]
	photoDir := filepath.Join(config.AppConfig.ImagesDirectory, fmt.Sprintf("%d", pregnancyID))

	// Create photo directory if it doesn't exist
	if len(files) > 0 {
		if err := os.MkdirAll(photoDir, 0755); err != nil {
			http.Error(w, "Failed to create photo directory", http.StatusInternalServerError)
			return
		}
	}

	// Process each uploaded photo
	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		// Generate unique filename
		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("%d_%d_%d%s", updateID, time.Now().Unix(), i, ext)
		fullPath := filepath.Join(photoDir, filename)

		// Create the file
		dst, err := os.Create(fullPath)
		if err != nil {
			continue
		}
		defer dst.Close()

		// Copy the uploaded file
		written, err := io.Copy(dst, file)
		if err != nil {
			continue
		}

		// Insert photo record
		_, err = db.GetDB().Exec(`
			INSERT INTO update_photos (update_id, filename, original_filename, file_size, sort_order)
			VALUES (?, ?, ?, ?, ?)`,
			updateID, filename, fileHeader.Filename, written, i)
		if err != nil {
			// Log error but continue processing other photos
			fmt.Printf("Failed to insert photo record: %v\n", err)
		}
	}

	// Handle video uploads
	videoFiles := r.MultipartForm.File["videos"]
	videoDir := filepath.Join(config.AppConfig.VideosDirectory, fmt.Sprintf("%d", pregnancyID))

	// Create video directory if it doesn't exist
	if len(videoFiles) > 0 {
		if err := os.MkdirAll(videoDir, 0755); err != nil {
			http.Error(w, "Failed to create video directory", http.StatusInternalServerError)
			return
		}
	}

	// Process each uploaded video
	for i, fileHeader := range videoFiles {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		// Validate video file type
		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if ext != ".mp4" && ext != ".mov" {
			continue // Skip unsupported video formats
		}

		// Generate unique filename
		filename := fmt.Sprintf("%d_%d_%d%s", updateID, time.Now().Unix(), i, ext)
		fullPath := filepath.Join(videoDir, filename)

		// Create the file
		dst, err := os.Create(fullPath)
		if err != nil {
			continue
		}
		defer dst.Close()

		// Copy the uploaded file
		written, err := io.Copy(dst, file)
		if err != nil {
			continue
		}

		// Insert video record (we'll reuse the update_photos table for now, adding a video table later if needed)
		_, err = db.GetDB().Exec(`
			INSERT INTO update_photos (update_id, filename, original_filename, file_size, sort_order)
			VALUES (?, ?, ?, ?, ?)`,
			updateID, filename, fileHeader.Filename, written, len(files)+i)
		if err != nil {
			// Log error but continue processing other videos
			fmt.Printf("Failed to insert video record: %v\n", err)
		}
	}

	// If update is shared, create an event
	if req.IsShared {
		// Get user details
		var userName string
		err = db.GetDB().QueryRow(`SELECT name FROM users WHERE id = ?`, userID).Scan(&userName)
		if err == nil && userName != "" {
			eventTitle := fmt.Sprintf("%s shared an update", userName)
			eventDescription := req.Title
			if req.Content != nil && *req.Content != "" {
				eventDescription = *req.Content
			}

			CreateUpdateSharedEvent(pregnancyID, userID, eventTitle, eventDescription, weekNumber)
		}
	}

	// Return the created update
	var update models.PregnancyUpdate
	err = db.GetDB().QueryRow(`
		SELECT id, pregnancy_id, week_number, title, content, update_type, appointment_type, is_shared, shared_at, update_date, created_at, updated_at
		FROM pregnancy_updates WHERE id = ?`, updateID).Scan(
		&update.ID, &update.PregnancyID, &update.WeekNumber, &update.Title, &update.Content,
		&update.UpdateType, &update.AppointmentType, &update.IsShared, &update.SharedAt, &update.UpdateDate,
		&update.CreatedAt, &update.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to fetch created update", http.StatusInternalServerError)
		return
	}

	// Get photos for the update
	rows, _ := db.GetDB().Query(`
		SELECT id, update_id, filename, original_filename, file_size, caption, sort_order, created_at
		FROM update_photos WHERE update_id = ? ORDER BY sort_order`, updateID)
	defer rows.Close()

	for rows.Next() {
		var photo models.UpdatePhoto
		rows.Scan(&photo.ID, &photo.UpdateID, &photo.Filename, &photo.OriginalFilename,
			&photo.FileSize, &photo.Caption, &photo.SortOrder, &photo.CreatedAt)
		update.Photos = append(update.Photos, photo)
	}

	// Send email notification if update is shared
	if update.IsShared {
		go func() {
			// Send email notifications in background to avoid blocking the response
			emailService, err := email.NewEmailService()
			if err != nil {
				log.Printf("Failed to initialize email service: %v", err)
				return
			}

			// Get pregnancy details
			var pregnancy models.Pregnancy
			query := `SELECT id, user_id, due_date, conception_date, baby_name, partner_name, 
					 partner_email, share_id, is_active, created_at 
					 FROM pregnancies WHERE id = ?`
			err = db.GetDB().QueryRow(query, pregnancyID).Scan(
				&pregnancy.ID, &pregnancy.UserID, &pregnancy.DueDate, &pregnancy.ConceptionDate,
				&pregnancy.BabyName, &pregnancy.PartnerName, &pregnancy.PartnerEmail,
				&pregnancy.ShareID, &pregnancy.IsActive, &pregnancy.CreatedAt,
			)
			if err != nil {
				log.Printf("Failed to get pregnancy for email notification: %v", err)
				return
			}

			// Send notification
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			err = emailService.SendUpdateNotification(ctx, &update, &pregnancy)
			if err != nil {
				log.Printf("Failed to send update notification: %v", err)
			} else {
				log.Printf("Update notification sent for pregnancy %d", pregnancyID)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(update)
}

// GetUpdatesHandler returns pregnancy updates for the current user
func GetUpdatesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Check if user is viewing as villager
	viewAsVillager := r.URL.Query().Get("view") == "villager"
	pregnancyIDStr := r.URL.Query().Get("pregnancy_id")

	var pregnancyID int
	var isOwner bool

	if pregnancyIDStr != "" {
		// Viewing specific pregnancy (villager view)
		pregnancyID, _ = strconv.Atoi(pregnancyIDStr)
		
		// Verify user has access
		var hasAccess bool
		err := db.GetDB().QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM pregnancies WHERE id = ? AND user_id = ?
				UNION
				SELECT 1 FROM village_members WHERE pregnancy_id = ? AND user_id = ?
			)`, pregnancyID, userID, pregnancyID, userID).Scan(&hasAccess)
		
		if err != nil || !hasAccess {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}

		// Check if owner
		db.GetDB().QueryRow(`SELECT user_id = ? FROM pregnancies WHERE id = ?`, userID, pregnancyID).Scan(&isOwner)
	} else {
		// Get user's own pregnancy
		err := db.GetDB().QueryRow(`
			SELECT id FROM pregnancies 
			WHERE user_id = ? AND is_active = TRUE
			ORDER BY created_at DESC LIMIT 1`,
			userID).Scan(&pregnancyID)
		if err != nil {
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		isOwner = true
	}

	// Build query based on access
	query := `
		SELECT id, pregnancy_id, week_number, title, content, update_type, 
		       appointment_type, is_shared, shared_at, update_date, created_at, updated_at
		FROM pregnancy_updates 
		WHERE pregnancy_id = ?`
	
	// If not owner or viewing as villager, only show shared updates
	if !isOwner || viewAsVillager {
		query += " AND is_shared = TRUE"
	}
	
	query += " ORDER BY COALESCE(update_date, created_at) DESC"

	rows, err := db.GetDB().Query(query, pregnancyID)
	if err != nil {
		http.Error(w, "Failed to fetch updates", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	updates := []models.PregnancyUpdate{}
	for rows.Next() {
		var update models.PregnancyUpdate
		err := rows.Scan(&update.ID, &update.PregnancyID, &update.WeekNumber, &update.Title,
			&update.Content, &update.UpdateType, &update.AppointmentType, &update.IsShared,
			&update.SharedAt, &update.UpdateDate, &update.CreatedAt, &update.UpdatedAt)
		if err != nil {
			continue
		}

		// Get photos for each update
		photoRows, _ := db.GetDB().Query(`
			SELECT id, update_id, filename, original_filename, file_size, caption, sort_order, created_at
			FROM update_photos WHERE update_id = ? ORDER BY sort_order`, update.ID)
		
		for photoRows.Next() {
			var photo models.UpdatePhoto
			photoRows.Scan(&photo.ID, &photo.UpdateID, &photo.Filename, &photo.OriginalFilename,
				&photo.FileSize, &photo.Caption, &photo.SortOrder, &photo.CreatedAt)
			update.Photos = append(update.Photos, photo)
		}
		photoRows.Close()

		updates = append(updates, update)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updates)
}

// UpdateShareStatusHandler updates the sharing status of an update
func UpdateShareStatusHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get update ID from URL
	urlParts := strings.Split(r.URL.Path, "/")
	if len(urlParts) < 2 {
		http.Error(w, "Invalid update ID", http.StatusBadRequest)
		return
	}

	updateID, err := strconv.Atoi(urlParts[len(urlParts)-2])
	if err != nil {
		http.Error(w, "Invalid update ID", http.StatusBadRequest)
		return
	}

	// Parse request
	var req struct {
		IsShared bool `json:"is_shared"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	// Verify ownership
	var pregnancyID int
	var currentlyShared bool
	err = db.GetDB().QueryRow(`
		SELECT p.id, pu.is_shared
		FROM pregnancy_updates pu
		JOIN pregnancies p ON p.id = pu.pregnancy_id
		WHERE pu.id = ? AND p.user_id = ?`,
		updateID, userID).Scan(&pregnancyID, &currentlyShared)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Update not found or access denied", http.StatusNotFound)
		return
	}

	// Update sharing status
	sharedAt := func() *time.Time {
		if req.IsShared {
			now := time.Now()
			return &now
		}
		return nil
	}()

	_, err = db.GetDB().Exec(`
		UPDATE pregnancy_updates 
		SET is_shared = ?, shared_at = ?
		WHERE id = ?`,
		req.IsShared, sharedAt, updateID)

	if err != nil {
		http.Error(w, "Failed to update sharing status", http.StatusInternalServerError)
		return
	}

	// Create event if newly shared
	if req.IsShared && !currentlyShared {
		var update models.PregnancyUpdate
		db.GetDB().QueryRow(`
			SELECT title, content FROM pregnancy_updates WHERE id = ?`,
			updateID).Scan(&update.Title, &update.Content)

		var userName string
		db.GetDB().QueryRow(`SELECT name FROM users WHERE id = ?`, userID).Scan(&userName)

		eventTitle := fmt.Sprintf("%s shared an update", userName)
		eventDescription := update.Title
		if update.Content != nil && *update.Content != "" {
			eventDescription = *update.Content
		}

		var weekNumber *int
		db.GetDB().QueryRow(`
			SELECT week_number FROM pregnancy_updates WHERE id = ?`,
			updateID).Scan(&weekNumber)

		CreateUpdateSharedEvent(pregnancyID, userID, eventTitle, eventDescription, weekNumber)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"is_shared": req.IsShared,
	})
}

// UpdateUpdateHandler handles updating an existing pregnancy update
func UpdateUpdateHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get update ID from URL
	urlParts := strings.Split(r.URL.Path, "/")
	if len(urlParts) < 3 {
		http.Error(w, "Invalid update ID", http.StatusBadRequest)
		return
	}

	updateID, err := strconv.Atoi(urlParts[len(urlParts)-1])
	if err != nil {
		http.Error(w, "Invalid update ID", http.StatusBadRequest)
		return
	}

	// Parse multipart form data for file uploads
	err = r.ParseMultipartForm(100 << 20) // 100 MB max for videos
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Parse JSON data from form field
	var req CreateUpdateRequest
	jsonData := r.FormValue("data")
	if jsonData != "" {
		if err := json.Unmarshal([]byte(jsonData), &req); err != nil {
			http.Error(w, "Invalid request data", http.StatusBadRequest)
			return
		}
	}

	// Validate required fields
	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Verify ownership and get conception date
	var pregnancyID int
	var conceptionDate *time.Time
	err = db.GetDB().QueryRow(`
		SELECT p.id, p.conception_date
		FROM pregnancy_updates pu
		JOIN pregnancies p ON p.id = pu.pregnancy_id
		WHERE pu.id = ? AND p.user_id = ?`,
		updateID, userID).Scan(&pregnancyID, &conceptionDate)
	
	if err == sql.ErrNoRows {
		http.Error(w, "Update not found or access denied", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Parse update date or use current UTC time
	var updateDate time.Time
	if req.Date != nil && *req.Date != "" {
		parsedDate, err := time.Parse(time.RFC3339, *req.Date)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			return
		}
		updateDate = parsedDate.UTC()
	} else {
		updateDate = time.Now().UTC()
	}

	// Calculate week number based on conception date
	var weekNumber *int
	if conceptionDate != nil {
		// Calculate gestational weeks from LMP (conception + 14 days)
		daysSinceConception := int(updateDate.Sub(*conceptionDate).Hours() / 24)
		// Add 14 days to account for LMP offset (gestational age calculation)
		calculatedWeek := (daysSinceConception+14)/7 + 1
		if calculatedWeek > 0 {
			weekNumber = &calculatedWeek
		}
	}

	// Update the existing update
	_, err = db.GetDB().Exec(`
		UPDATE pregnancy_updates 
		SET week_number = ?, title = ?, content = ?, update_type = ?, appointment_type = ?, 
		    is_shared = ?, shared_at = ?, update_date = ?, updated_at = ?
		WHERE id = ?`,
		weekNumber, req.Title, req.Content, req.UpdateType, req.AppointmentType, req.IsShared,
		func() *time.Time {
			if req.IsShared {
				now := time.Now()
				return &now
			}
			return nil
		}(), &updateDate, time.Now(), updateID)

	if err != nil {
		http.Error(w, "Failed to update", http.StatusInternalServerError)
		return
	}

	// Handle new photo uploads (append to existing photos)
	files := r.MultipartForm.File["photos"]
	if len(files) > 0 {
		photoDir := filepath.Join(config.AppConfig.ImagesDirectory, fmt.Sprintf("%d", pregnancyID))
		
		// Create photo directory if it doesn't exist
		if err := os.MkdirAll(photoDir, 0755); err != nil {
			http.Error(w, "Failed to create photo directory", http.StatusInternalServerError)
			return
		}

		// Get current max sort order
		var maxSortOrder int
		db.GetDB().QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM update_photos WHERE update_id = ?`, updateID).Scan(&maxSortOrder)

		// Process each uploaded photo
		for i, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}
			defer file.Close()

			// Generate unique filename
			ext := filepath.Ext(fileHeader.Filename)
			filename := fmt.Sprintf("%d_%d_%d%s", updateID, time.Now().Unix(), i, ext)
			fullPath := filepath.Join(photoDir, filename)

			// Create the file
			dst, err := os.Create(fullPath)
			if err != nil {
				continue
			}
			defer dst.Close()

			// Copy the uploaded file
			written, err := io.Copy(dst, file)
			if err != nil {
				continue
			}

			// Insert photo record
			_, err = db.GetDB().Exec(`
				INSERT INTO update_photos (update_id, filename, original_filename, file_size, sort_order)
				VALUES (?, ?, ?, ?, ?)`,
				updateID, filename, fileHeader.Filename, written, maxSortOrder+i+1)
			if err != nil {
				fmt.Printf("Failed to insert photo record: %v\n", err)
			}
		}
	}

	// Handle new video uploads (append to existing videos)
	videoFiles := r.MultipartForm.File["videos"]
	if len(videoFiles) > 0 {
		videoDir := filepath.Join(config.AppConfig.VideosDirectory, fmt.Sprintf("%d", pregnancyID))
		
		// Create video directory if it doesn't exist
		if err := os.MkdirAll(videoDir, 0755); err != nil {
			http.Error(w, "Failed to create video directory", http.StatusInternalServerError)
			return
		}

		// Get current max sort order
		var maxSortOrder int
		db.GetDB().QueryRow(`SELECT COALESCE(MAX(sort_order), -1) FROM update_photos WHERE update_id = ?`, updateID).Scan(&maxSortOrder)

		// Process each uploaded video
		for i, fileHeader := range videoFiles {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}
			defer file.Close()

			// Validate video file type
			ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
			if ext != ".mp4" && ext != ".mov" {
				continue // Skip unsupported video formats
			}

			// Generate unique filename
			filename := fmt.Sprintf("%d_%d_%d%s", updateID, time.Now().Unix(), i, ext)
			fullPath := filepath.Join(videoDir, filename)

			// Create the file
			dst, err := os.Create(fullPath)
			if err != nil {
				continue
			}
			defer dst.Close()

			// Copy the uploaded file
			written, err := io.Copy(dst, file)
			if err != nil {
				continue
			}

			// Insert video record
			_, err = db.GetDB().Exec(`
				INSERT INTO update_photos (update_id, filename, original_filename, file_size, sort_order)
				VALUES (?, ?, ?, ?, ?)`,
				updateID, filename, fileHeader.Filename, written, maxSortOrder+len(files)+i+1)
			if err != nil {
				fmt.Printf("Failed to insert video record: %v\n", err)
			}
		}
	}

	// Return the updated update
	var update models.PregnancyUpdate
	err = db.GetDB().QueryRow(`
		SELECT id, pregnancy_id, week_number, title, content, update_type, appointment_type, is_shared, shared_at, update_date, created_at, updated_at
		FROM pregnancy_updates WHERE id = ?`, updateID).Scan(
		&update.ID, &update.PregnancyID, &update.WeekNumber, &update.Title, &update.Content,
		&update.UpdateType, &update.AppointmentType, &update.IsShared, &update.SharedAt, &update.UpdateDate,
		&update.CreatedAt, &update.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to fetch updated update", http.StatusInternalServerError)
		return
	}

	// Get photos for the update
	rows, _ := db.GetDB().Query(`
		SELECT id, update_id, filename, original_filename, file_size, caption, sort_order, created_at
		FROM update_photos WHERE update_id = ? ORDER BY sort_order`, updateID)
	defer rows.Close()

	for rows.Next() {
		var photo models.UpdatePhoto
		rows.Scan(&photo.ID, &photo.UpdateID, &photo.Filename, &photo.OriginalFilename,
			&photo.FileSize, &photo.Caption, &photo.SortOrder, &photo.CreatedAt)
		update.Photos = append(update.Photos, photo)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(update)
}