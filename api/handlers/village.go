package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
)

type CreateVillageMemberRequest struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Relationship string `json:"relationship"`
	IsTold       bool   `json:"is_told"`
}

type CreateVillageMembersBulkRequest struct {
	Name         string   `json:"name"`
	Emails       []string `json:"emails"`
	Relationship string   `json:"relationship"`
	IsTold       bool     `json:"is_told"`
}

type CreateVillageMembersBulkResponse struct {
	Members []*models.VillageMember `json:"members"`
}

// CreateVillageMemberHandler handles creating a new village member
func CreateVillageMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateVillageMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim whitespace
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Relationship = strings.TrimSpace(req.Relationship)

	// Validate required fields
	if req.Name == "" || req.Email == "" || req.Relationship == "" {
		http.Error(w, "Name, email, and relationship are required", http.StatusBadRequest)
		return
	}

	// Get user's pregnancy
	pregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting pregnancy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if email already exists for this pregnancy
	existingMember, err := GetVillageMemberByEmail(pregnancy.ID, req.Email)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Database error checking existing member: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingMember != nil {
		http.Error(w, "This email is already in your village", http.StatusConflict)
		return
	}

	// Create the village member with event
	member, err := CreateVillageMemberWithEvent(pregnancy.ID, req.Name, req.Email, req.Relationship, req.IsTold, false)
	if err != nil {
		log.Printf("Failed to create village member: %v", err)
		http.Error(w, "Failed to create village member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(member)
}

// CreateVillageMembersBulkHandler handles creating multiple village members from comma-separated emails
func CreateVillageMembersBulkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context (set by auth middleware)
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateVillageMembersBulkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim whitespace from name and relationship
	req.Name = strings.TrimSpace(req.Name)
	req.Relationship = strings.TrimSpace(req.Relationship)

	// Validate required fields
	if req.Name == "" || len(req.Emails) == 0 || req.Relationship == "" {
		http.Error(w, "Name, emails, and relationship are required", http.StatusBadRequest)
		return
	}

	// Trim whitespace from emails
	for i, email := range req.Emails {
		req.Emails[i] = strings.TrimSpace(email)
		if req.Emails[i] == "" {
			http.Error(w, "Empty email not allowed", http.StatusBadRequest)
			return
		}
	}

	// Get user's pregnancy
	pregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting pregnancy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if any emails already exist for this pregnancy
	for _, email := range req.Emails {
		existingMember, err := GetVillageMemberByEmail(pregnancy.ID, email)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Database error checking existing member: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if existingMember != nil {
			http.Error(w, fmt.Sprintf("Email %s is already in your village", email), http.StatusConflict)
			return
		}
	}

	// Create village members for each email
	var members []*models.VillageMember
	for i, email := range req.Emails {
		// For multiple emails, append number to name (e.g., "John & Jane" becomes "John & Jane (1)", "John & Jane (2)")
		memberName := req.Name
		if len(req.Emails) > 1 {
			memberName = fmt.Sprintf("%s (%d)", req.Name, i+1)
		}

		member, err := CreateVillageMemberWithEvent(pregnancy.ID, memberName, email, req.Relationship, req.IsTold, false)
		if err != nil {
			log.Printf("Failed to create village member: %v", err)
			http.Error(w, "Failed to create village member", http.StatusInternalServerError)
			return
		}
		members = append(members, member)
	}

	response := CreateVillageMembersBulkResponse{
		Members: members,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetVillageMembersHandler returns all village members for the current user's pregnancy
func GetVillageMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's pregnancy
	pregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting pregnancy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	members, err := GetVillageMembersByPregnancyID(pregnancy.ID)
	if err != nil {
		log.Printf("Database error getting village members: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(members)
}

// DeleteVillageMemberHandler deletes a village member
func DeleteVillageMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract member ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/village-members/")
	memberID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	// Get user's pregnancy
	pregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting pregnancy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify the member belongs to this user's pregnancy
	member, err := GetVillageMemberByID(memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Village member not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting village member: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if member.PregnancyID != pregnancy.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Delete the member
	err = DeleteVillageMember(memberID)
	if err != nil {
		log.Printf("Failed to delete village member: %v", err)
		http.Error(w, "Failed to delete village member", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateVillageMemberHandler updates a village member's status
func UpdateVillageMemberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract member ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/village-members/")
	memberID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid member ID", http.StatusBadRequest)
		return
	}

	var req struct {
		IsTold bool `json:"is_told"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user's pregnancy
	pregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting pregnancy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Verify the member belongs to this user's pregnancy
	member, err := GetVillageMemberByID(memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Village member not found", http.StatusNotFound)
			return
		}
		log.Printf("Database error getting village member: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if member.PregnancyID != pregnancy.ID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Update the member's status
	updatedMember, err := UpdateVillageMemberStatusWithEvent(memberID, req.IsTold, claims.UserID)
	if err != nil {
		log.Printf("Failed to update village member: %v", err)
		http.Error(w, "Failed to update village member", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedMember)
}

// Database functions

func CreateVillageMember(pregnancyID int, name, email, relationship string, isTold bool) (*models.VillageMember, error) {
	query := `
		INSERT INTO village_members (pregnancy_id, name, email, relationship, is_told)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, pregnancy_id, name, email, relationship, is_told, told_date, is_subscribed, unsubscribe_token, created_at, updated_at
	`

	var member models.VillageMember
	err := db.GetDB().QueryRow(query, pregnancyID, name, email, relationship, isTold).Scan(
		&member.ID,
		&member.PregnancyID,
		&member.Name,
		&member.Email,
		&member.Relationship,
		&member.IsTold,
		&member.ToldDate,
		&member.IsSubscribed,
		&member.UnsubscribeToken,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &member, nil
}

// CreateVillageMemberWithEvent creates a village member and corresponding event
func CreateVillageMemberWithEvent(pregnancyID int, name, email, relationship string, isTold bool, isFromInvite bool) (*models.VillageMember, error) {
	// Create the village member first
	member, err := CreateVillageMember(pregnancyID, name, email, relationship, isTold)
	if err != nil {
		return nil, err
	}

	// Get current week for the pregnancy
	pregnancy, err := GetPregnancyByID(pregnancyID)
	if err != nil {
		log.Printf("Could not get pregnancy for event creation: %v", err)
		return member, nil // Don't fail member creation if event fails
	}

	weekNumber := pregnancy.GetCurrentWeek()

	// Create appropriate event based on how they joined
	if isFromInvite {
		if err := CreateVillagerJoinedEvent(pregnancyID, name, relationship, &weekNumber); err != nil {
			log.Printf("Failed to create villager joined event: %v", err)
		}
	} else {
		// For manually added villagers, create a "villager added" event
		if err := CreateVillagerAddedEvent(pregnancyID, name, relationship, &weekNumber); err != nil {
			log.Printf("Failed to create villager added event: %v", err)
		}
	}

	return member, nil
}

func GetVillageMembersByPregnancyID(pregnancyID int) ([]*models.VillageMember, error) {
	query := `
		SELECT id, pregnancy_id, name, email, relationship, is_told, told_date, is_subscribed, unsubscribe_token, created_at, updated_at
		FROM village_members 
		WHERE pregnancy_id = ?
		ORDER BY created_at ASC
	`

	rows, err := db.GetDB().Query(query, pregnancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*models.VillageMember
	for rows.Next() {
		var member models.VillageMember
		err := rows.Scan(
			&member.ID,
			&member.PregnancyID,
			&member.Name,
			&member.Email,
			&member.Relationship,
			&member.IsTold,
			&member.ToldDate,
			&member.IsSubscribed,
			&member.UnsubscribeToken,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		members = append(members, &member)
	}

	return members, nil
}

func GetVillageMemberByEmail(pregnancyID int, email string) (*models.VillageMember, error) {
	query := `
		SELECT id, pregnancy_id, name, email, relationship, is_told, told_date, is_subscribed, unsubscribe_token, created_at, updated_at
		FROM village_members 
		WHERE pregnancy_id = ? AND email = ?
		LIMIT 1
	`

	var member models.VillageMember
	err := db.GetDB().QueryRow(query, pregnancyID, email).Scan(
		&member.ID,
		&member.PregnancyID,
		&member.Name,
		&member.Email,
		&member.Relationship,
		&member.IsTold,
		&member.ToldDate,
		&member.IsSubscribed,
		&member.UnsubscribeToken,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func GetVillageMemberByID(memberID int) (*models.VillageMember, error) {
	query := `
		SELECT id, pregnancy_id, name, email, relationship, is_told, told_date, is_subscribed, unsubscribe_token, created_at, updated_at
		FROM village_members 
		WHERE id = ?
		LIMIT 1
	`

	var member models.VillageMember
	err := db.GetDB().QueryRow(query, memberID).Scan(
		&member.ID,
		&member.PregnancyID,
		&member.Name,
		&member.Email,
		&member.Relationship,
		&member.IsTold,
		&member.ToldDate,
		&member.IsSubscribed,
		&member.UnsubscribeToken,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &member, nil
}

func UpdateVillageMemberStatus(memberID int, isTold bool) (*models.VillageMember, error) {
	query := `
		UPDATE village_members 
		SET is_told = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING id, pregnancy_id, name, email, relationship, is_told, told_date, is_subscribed, unsubscribe_token, created_at, updated_at
	`

	var member models.VillageMember
	err := db.GetDB().QueryRow(query, isTold, memberID).Scan(
		&member.ID,
		&member.PregnancyID,
		&member.Name,
		&member.Email,
		&member.Relationship,
		&member.IsTold,
		&member.ToldDate,
		&member.IsSubscribed,
		&member.UnsubscribeToken,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &member, nil
}

// UpdateVillageMemberStatusWithEvent updates villager status and creates event
func UpdateVillageMemberStatusWithEvent(memberID int, isTold bool, userID int) (*models.VillageMember, error) {
	// Get the member before updating to check if status actually changed
	originalMember, err := GetVillageMemberByID(memberID)
	if err != nil {
		return nil, err
	}

	// Update the member status
	member, err := UpdateVillageMemberStatus(memberID, isTold)
	if err != nil {
		return nil, err
	}

	// Create event only if status changed from false to true (newly told)
	if !originalMember.IsTold && isTold {
		// Get current week for the pregnancy
		pregnancy, err := GetPregnancyByID(member.PregnancyID)
		if err != nil {
			log.Printf("Could not get pregnancy for event creation: %v", err)
			return member, nil // Don't fail update if event fails
		}

		weekNumber := pregnancy.GetCurrentWeek()
		if err := CreateVillagerToldEvent(member.PregnancyID, member.Name, userID, &weekNumber); err != nil {
			log.Printf("Failed to create villager told event: %v", err)
		}
	}

	return member, nil
}

func DeleteVillageMember(memberID int) error {
	query := `DELETE FROM village_members WHERE id = ?`
	_, err := db.GetDB().Exec(query, memberID)
	return err
}