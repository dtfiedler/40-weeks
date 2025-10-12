package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"simple-go/api/db"
	"simple-go/api/models"
	"simple-go/api/services/email"
)

// PublicTimelineItem represents a public timeline item for village members
type PublicTimelineItem struct {
	ID          int                  `json:"id"`
	Title       string               `json:"title"`
	Description *string              `json:"description"`
	WeekNumber  *int                 `json:"week_number"`
	UpdateDate  string               `json:"update_date"`
	CreatedBy   string               `json:"created_by"`
	Photos      []models.UpdatePhoto `json:"photos,omitempty"`
	PregnancyID int                  `json:"pregnancy_id"`
}

// VerifyAccessRequest represents a request to verify email access
type VerifyAccessRequest struct {
	Email string `json:"email"`
}

// VerifyAccessResponse represents the response for access verification
type VerifyAccessResponse struct {
	HasAccess bool `json:"has_access"`
}

// AccessRequest represents a request for timeline access
type AccessRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	Message      string `json:"message"`
}

// PublicTimelineHandler returns shared updates for a pregnancy via share ID
func PublicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract share ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/timeline/")
	shareID := path
	
	if shareID == "" {
		http.Error(w, "Invalid share ID", http.StatusBadRequest)
		return
	}

	// Require email parameter for access verification
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email parameter required", http.StatusBadRequest)
		return
	}

	// Get pregnancy by share ID
	pregnancy, err := GetPregnancyByShareID(shareID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Pregnancy not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify email has access to this timeline
	hasAccess, err := verifyEmailAccess(pregnancy, email)
	if err != nil {
		log.Printf("Error verifying email access: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Parse limit and offset
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get public timeline items (only shared updates)
	items, err := getPublicTimelineItems(pregnancy.ID, limit, offset)
	if err != nil {
		log.Printf("Failed to get public timeline items: %v", err)
		http.Error(w, "Failed to retrieve timeline", http.StatusInternalServerError)
		return
	}

	// Get pregnancy info for context
	var userName string
	err = db.GetDB().QueryRow(`SELECT name FROM users WHERE id = ?`, pregnancy.UserID).Scan(&userName)
	if err != nil {
		log.Printf("Failed to get user name: %v", err)
		userName = "Parent"
	}

	parentNames := strings.Fields(strings.TrimSpace(userName))[0]
	if pregnancy.PartnerName != nil && *pregnancy.PartnerName != "" {
		parentNames = parentNames + " & " + strings.Fields(strings.TrimSpace(*pregnancy.PartnerName))[0]
	}

	babyName := "Baby"
	if pregnancy.BabyName != nil && *pregnancy.BabyName != "" {
		babyName = *pregnancy.BabyName
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updates":      items,
		"total":        len(items),
		"pregnancy": map[string]interface{}{
			"parent_names":  parentNames,
			"baby_name":     babyName,
			"due_date":      pregnancy.DueDate.Format("2006-01-02"),
			"current_week":  pregnancy.GetCurrentWeek(),
		},
	})
}

// VerifyTimelineAccessHandler checks if an email has access to view the timeline
func VerifyTimelineAccessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract share ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/timeline/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "verify-access" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	shareID := parts[0]

	if shareID == "" {
		http.Error(w, "Invalid share ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req VerifyAccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Get pregnancy by share ID
	pregnancy, err := GetPregnancyByShareID(shareID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Pregnancy not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if email is the pregnancy owner
	var ownerEmail string
	err = db.GetDB().QueryRow(`
		SELECT u.email 
		FROM users u 
		WHERE u.id = ?
	`, pregnancy.UserID).Scan(&ownerEmail)
	
	if err != nil {
		log.Printf("Error getting pregnancy owner email: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if user is pregnancy owner or partner
	isOwner := strings.EqualFold(req.Email, ownerEmail)
	isPartner := pregnancy.PartnerEmail != nil && strings.EqualFold(req.Email, *pregnancy.PartnerEmail)

	if isOwner || isPartner {
		response := VerifyAccessResponse{
			HasAccess: true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if email belongs to a village member
	var count int
	err = db.GetDB().QueryRow(`
		SELECT COUNT(*) 
		FROM village_members 
		WHERE pregnancy_id = ? AND LOWER(email) = LOWER(?)
	`, pregnancy.ID, req.Email).Scan(&count)
	
	if err != nil {
		log.Printf("Error checking village member access: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := VerifyAccessResponse{
		HasAccess: count > 0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RequestTimelineAccessHandler handles access requests from non-village members
func RequestTimelineAccessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract share ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/timeline/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[1] != "request-access" {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	shareID := parts[0]

	if shareID == "" {
		http.Error(w, "Invalid share ID", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req AccessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Name == "" || req.Relationship == "" {
		http.Error(w, "Email, name, and relationship are required", http.StatusBadRequest)
		return
	}

	// Get pregnancy by share ID
	pregnancy, err := GetPregnancyByShareID(shareID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Pregnancy not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if they're already in the village
	var count int
	err = db.GetDB().QueryRow(`
		SELECT COUNT(*) 
		FROM village_members 
		WHERE pregnancy_id = ? AND LOWER(email) = LOWER(?)
	`, pregnancy.ID, req.Email).Scan(&count)
	
	if err != nil {
		log.Printf("Error checking existing village member: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "You are already a member of this village", http.StatusBadRequest)
		return
	}

	// Check if there's already a pending request from this email
	var existingRequestCount int
	err = db.GetDB().QueryRow(`
		SELECT COUNT(*) 
		FROM access_requests 
		WHERE pregnancy_id = ? AND LOWER(email) = LOWER(?) AND status = 'pending'
	`, pregnancy.ID, req.Email).Scan(&existingRequestCount)
	
	if err != nil {
		log.Printf("Error checking existing access request: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingRequestCount > 0 {
		http.Error(w, "You already have a pending access request for this pregnancy", http.StatusBadRequest)
		return
	}

	// Store the access request in the database
	_, err = db.GetDB().Exec(`
		INSERT INTO access_requests (pregnancy_id, email, name, relationship, message, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'pending', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, pregnancy.ID, req.Email, req.Name, req.Relationship, req.Message)

	if err != nil {
		log.Printf("Error storing access request: %v", err)
		http.Error(w, "Failed to store access request", http.StatusInternalServerError)
		return
	}

	log.Printf("Access request stored for pregnancy %d: %s (%s) wants to join as %s", 
		pregnancy.ID, req.Name, req.Email, req.Relationship)

	// Send email notification to pregnancy owner
	emailService, err := email.NewEmailService()
	if err != nil {
		log.Printf("Error creating email service: %v", err)
	} else {
		// Send notification in the background to avoid blocking the response
		go func() {
			ctx := context.Background()
			err := emailService.SendAccessRequestNotification(ctx, pregnancy.ID, req.Name, req.Email, req.Relationship, req.Message)
			if err != nil {
				log.Printf("Error sending access request notification: %v", err)
			} else {
				log.Printf("Access request notification sent for pregnancy %d", pregnancy.ID)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// verifyEmailAccess checks if an email has access to view the timeline
func verifyEmailAccess(pregnancy *models.Pregnancy, email string) (bool, error) {
	// Check if email is the pregnancy owner
	var ownerEmail string
	err := db.GetDB().QueryRow(`
		SELECT u.email 
		FROM users u 
		WHERE u.id = ?
	`, pregnancy.UserID).Scan(&ownerEmail)
	
	if err != nil {
		return false, err
	}

	// Check if user is pregnancy owner or partner
	isOwner := strings.EqualFold(email, ownerEmail)
	isPartner := pregnancy.PartnerEmail != nil && strings.EqualFold(email, *pregnancy.PartnerEmail)

	if isOwner || isPartner {
		return true, nil
	}

	// Check if email belongs to a village member
	var count int
	err = db.GetDB().QueryRow(`
		SELECT COUNT(*) 
		FROM village_members 
		WHERE pregnancy_id = ? AND LOWER(email) = LOWER(?)
	`, pregnancy.ID, email).Scan(&count)
	
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// getPublicTimelineItems fetches only shared updates for public viewing
func getPublicTimelineItems(pregnancyID, limit, offset int) ([]PublicTimelineItem, error) {
	query := `
	SELECT 
		pu.id,
		pu.title,
		pu.content as description,
		pu.week_number,
		COALESCE(pu.update_date, pu.created_at) as update_date,
		u.name as created_by
	FROM pregnancy_updates pu
	JOIN pregnancies p ON p.id = pu.pregnancy_id
	JOIN users u ON u.id = p.user_id
	WHERE pu.pregnancy_id = ? AND pu.is_shared = TRUE
	ORDER BY update_date DESC, pu.id DESC
	LIMIT ? OFFSET ?`

	rows, err := db.GetDB().Query(query, pregnancyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PublicTimelineItem
	for rows.Next() {
		var item PublicTimelineItem
		var createdBy string
		var updateDateStr string
		
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.WeekNumber,
			&updateDateStr,
			&createdBy,
		)
		if err != nil {
			continue
		}

		// Parse the datetime from SQLite and format as ISO string
		if updateDateStr != "" {
			// SQLite datetime format is typically "2006-01-02 15:04:05" or "2006-01-02T15:04:05Z"
			var parsedTime time.Time
			
			// Try common SQLite datetime formats
			formats := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05",
				"2006-01-02",
			}
			
			for _, format := range formats {
				if parsed, err := time.Parse(format, updateDateStr); err == nil {
					parsedTime = parsed
					break
				}
			}
			
			// If we couldn't parse it, use the original string
			if !parsedTime.IsZero() {
				// Ensure it's UTC and format as ISO string
				item.UpdateDate = parsedTime.UTC().Format(time.RFC3339)
			} else {
				item.UpdateDate = updateDateStr
			}
		}

		item.CreatedBy = createdBy
		item.PregnancyID = pregnancyID

		// Get photos/videos for this update
		photos, _ := getUpdatePhotos(item.ID)
		item.Photos = photos

		items = append(items, item)
	}

	return items, nil
}
