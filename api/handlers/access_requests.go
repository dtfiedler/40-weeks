package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
	"simple-go/api/services/email"
)

// AccessRequestRecord represents an access request from the database
type AccessRequestRecord struct {
	ID           int    `json:"id"`
	PregnancyID  int    `json:"pregnancy_id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	Relationship string `json:"relationship"`
	Message      string `json:"message"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// GetAccessRequestsHandler returns all pending access requests for the current user's pregnancy
func GetAccessRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from JWT token (set by auth middleware)
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get the user's active pregnancy
	var pregnancyID int
	err := db.GetDB().QueryRow(`
		SELECT id FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE
		LIMIT 1
	`, userID).Scan(&pregnancyID)

	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]AccessRequestRecord{})
			return
		}
		log.Printf("Error getting pregnancy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get all pending access requests for this pregnancy
	rows, err := db.GetDB().Query(`
		SELECT id, pregnancy_id, email, name, relationship, message, status, created_at, updated_at
		FROM access_requests
		WHERE pregnancy_id = ? AND status = 'pending'
		ORDER BY created_at DESC
	`, pregnancyID)

	if err != nil {
		log.Printf("Error getting access requests: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var requests []AccessRequestRecord
	for rows.Next() {
		var req AccessRequestRecord
		err := rows.Scan(
			&req.ID,
			&req.PregnancyID,
			&req.Email,
			&req.Name,
			&req.Relationship,
			&req.Message,
			&req.Status,
			&req.CreatedAt,
			&req.UpdatedAt,
		)
		if err != nil {
			log.Printf("Error scanning access request: %v", err)
			continue
		}
		requests = append(requests, req)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

// ManageAccessRequestHandler approves or denies an access request
func ManageAccessRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from JWT token (set by auth middleware)
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Extract request ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/village-members/access-requests/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	action := parts[1] // should be "approve" or "deny"
	if action != "approve" && action != "deny" {
		http.Error(w, "Action must be 'approve' or 'deny'", http.StatusBadRequest)
		return
	}

	// Get the access request and verify it belongs to the user's pregnancy
	var req AccessRequestRecord
	var pregnancyUserID int
	err = db.GetDB().QueryRow(`
		SELECT ar.id, ar.pregnancy_id, ar.email, ar.name, ar.relationship, ar.message, ar.status, ar.created_at, ar.updated_at, p.user_id
		FROM access_requests ar
		JOIN pregnancies p ON p.id = ar.pregnancy_id
		WHERE ar.id = ? AND ar.status = 'pending'
	`, requestID).Scan(
		&req.ID,
		&req.PregnancyID,
		&req.Email,
		&req.Name,
		&req.Relationship,
		&req.Message,
		&req.Status,
		&req.CreatedAt,
		&req.UpdatedAt,
		&pregnancyUserID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Access request not found or already processed", http.StatusNotFound)
			return
		}
		log.Printf("Error getting access request: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify the request belongs to the authenticated user's pregnancy
	if pregnancyUserID != userID {
		http.Error(w, "Access denied: This request doesn't belong to your pregnancy", http.StatusForbidden)
		return
	}

	if action == "approve" {
		// Add the person to the village
		_, err = db.GetDB().Exec(`
			INSERT INTO village_members (pregnancy_id, name, email, relationship, is_told, created_at, updated_at)
			VALUES (?, ?, ?, ?, TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, req.PregnancyID, req.Name, req.Email, req.Relationship)

		if err != nil {
			log.Printf("Error adding village member: %v", err)
			http.Error(w, "Failed to add village member", http.StatusInternalServerError)
			return
		}

		log.Printf("Access request approved: %s (%s) added to pregnancy %d village", req.Name, req.Email, req.PregnancyID)

		// Send welcome email to the newly approved member
		emailService, err := email.NewEmailService()
		if err != nil {
			log.Printf("Error creating email service: %v", err)
		} else {
			// Get full pregnancy details for the email
			var pregnancy models.Pregnancy
			err = db.GetDB().QueryRow(`
				SELECT id, user_id, share_id, baby_name, partner_name, partner_email, due_date, created_at, updated_at, is_active, cover_photo_filename
				FROM pregnancies
				WHERE id = ?
			`, req.PregnancyID).Scan(
				&pregnancy.ID,
				&pregnancy.UserID,
				&pregnancy.ShareID,
				&pregnancy.BabyName,
				&pregnancy.PartnerName,
				&pregnancy.PartnerEmail,
				&pregnancy.DueDate,
				&pregnancy.CreatedAt,
				&pregnancy.UpdatedAt,
				&pregnancy.IsActive,
				&pregnancy.CoverPhotoFilename,
			)

			if err != nil {
				log.Printf("Error getting pregnancy details for welcome email: %v", err)
			} else {
				// Create a village member struct from the access request
				villageMember := &models.VillageMember{
					PregnancyID:  req.PregnancyID,
					Name:         req.Name,
					Email:        req.Email,
					Relationship: req.Relationship,
				}

				// Send welcome email asynchronously
				go func() {
					ctx := context.Background()
					err := emailService.SendWelcomeEmail(ctx, villageMember, &pregnancy)
					if err != nil {
						log.Printf("Error sending welcome email to %s: %v", req.Email, err)
					} else {
						log.Printf("Welcome email sent to newly approved member: %s", req.Email)
					}
				}()
			}
		}
	} else {
		log.Printf("Access request denied: %s (%s) for pregnancy %d", req.Name, req.Email, req.PregnancyID)
	}

	// Delete the access request (approved or denied)
	_, err = db.GetDB().Exec(`DELETE FROM access_requests WHERE id = ?`, requestID)
	if err != nil {
		log.Printf("Error deleting access request: %v", err)
		// Don't fail the request if we can't delete - the action was successful
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"action": action,
		"message": action + "d successfully",
	})
}