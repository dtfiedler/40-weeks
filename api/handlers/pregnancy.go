package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
)

type CreatePregnancyRequest struct {
	DueDate      string  `json:"due_date"`
	PartnerName  *string `json:"partner_name"`
	PartnerEmail *string `json:"partner_email"`
	BabyName     *string `json:"baby_name"`
}

type PregnancyResponse struct {
	*models.Pregnancy
	CurrentWeek int `json:"current_week_calculated"`
}

type InviteHashResponse struct {
	Hash string `json:"hash"`
}

type PregnancyInviteInfo struct {
	ParentNames string `json:"parent_names"`
	BabyName    string `json:"baby_name"`
	DueDate     string `json:"due_date"`
}

type JoinVillageRequest struct {
	Name         string   `json:"name"`
	Emails       []string `json:"emails"`
	Relationship string   `json:"relationship"`
	IsTold       bool     `json:"is_told"`
}

// CreatePregnancyHandler handles creating a new pregnancy
func CreatePregnancyHandler(w http.ResponseWriter, r *http.Request) {
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

	var req CreatePregnancyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse due date
	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		http.Error(w, "Invalid due date format", http.StatusBadRequest)
		return
	}

	// Check if user already has an active pregnancy
	existingPregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil && err != sql.ErrNoRows {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if existingPregnancy != nil {
		http.Error(w, "You already have an active pregnancy", http.StatusConflict)
		return
	}

	// Create the pregnancy
	pregnancy, err := CreatePregnancy(claims.UserID, dueDate, req.PartnerName, req.PartnerEmail, req.BabyName)
	if err != nil {
		http.Error(w, "Failed to create pregnancy", http.StatusInternalServerError)
		return
	}

	// Return the created pregnancy with calculated week
	response := &PregnancyResponse{
		Pregnancy:   pregnancy,
		CurrentWeek: pregnancy.GetCurrentWeek(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPregnancyHandler returns the current user's active pregnancy
func GetPregnancyHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("GetPregnancyHandler called with method: %s, path: %s", r.Method, r.URL.Path)
	
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		log.Printf("No claims found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	log.Printf("Looking for pregnancy for user ID: %d", claims.UserID)
	pregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No pregnancy found for user ID: %d", claims.UserID)
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		log.Printf("Database error looking for pregnancy: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := &PregnancyResponse{
		Pregnancy:   pregnancy,
		CurrentWeek: pregnancy.GetCurrentWeek(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetInviteHashHandler returns the invite hash for the user's pregnancy
func GetInviteHashHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	pregnancy, err := GetActivePregnancyByUserID(claims.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No active pregnancy found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	hash := generatePregnancyHash(pregnancy.ID)
	response := &InviteHashResponse{Hash: hash}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPregnancyFromInviteHandler returns pregnancy info from invite hash
func GetPregnancyFromInviteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract hash from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/pregnancy/invite/")
	hash := path

	if hash == "" {
		http.Error(w, "Invalid invite hash", http.StatusBadRequest)
		return
	}

	pregnancyID, err := validatePregnancyHash(hash)
	if err != nil {
		http.Error(w, "Invalid or expired invite", http.StatusNotFound)
		return
	}

	pregnancy, err := GetPregnancyByID(pregnancyID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Pregnancy not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Get user info to build parent names
	user, err := GetUserByID(pregnancy.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	parentNames := user.Name
	if pregnancy.PartnerName != nil && *pregnancy.PartnerName != "" {
		parentNames = fmt.Sprintf("%s & %s", user.Name, *pregnancy.PartnerName)
	}

	babyName := "Baby"
	if pregnancy.BabyName != nil && *pregnancy.BabyName != "" {
		babyName = *pregnancy.BabyName
	}

	response := &PregnancyInviteInfo{
		ParentNames: parentNames,
		BabyName:    babyName,
		DueDate:     pregnancy.DueDate.Format("2006-01-02"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// JoinVillageFromInviteHandler allows someone to join a village via invite link
func JoinVillageFromInviteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract hash from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/pregnancy/join/")
	hash := path

	if hash == "" {
		http.Error(w, "Invalid invite hash", http.StatusBadRequest)
		return
	}

	pregnancyID, err := validatePregnancyHash(hash)
	if err != nil {
		http.Error(w, "Invalid or expired invite", http.StatusNotFound)
		return
	}

	var req JoinVillageRequest
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

	// Check if any emails already exist for this pregnancy
	for _, email := range req.Emails {
		existingMember, err := GetVillageMemberByEmail(pregnancyID, email)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Database error checking existing member: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if existingMember != nil {
			http.Error(w, fmt.Sprintf("Email %s is already in this village", email), http.StatusConflict)
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

		member, err := CreateVillageMember(pregnancyID, memberName, email, req.Relationship, req.IsTold)
		if err != nil {
			log.Printf("Failed to create village member: %v", err)
			http.Error(w, "Failed to create village member", http.StatusInternalServerError)
			return
		}
		members = append(members, member)
	}

	response := map[string]interface{}{
		"members": members,
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Database functions

func CreatePregnancy(userID int, dueDate time.Time, partnerName, partnerEmail, babyName *string) (*models.Pregnancy, error) {
	// Set default baby name if empty
	if babyName == nil || *babyName == "" {
		defaultName := "Baby"
		babyName = &defaultName
	}

	query := `
		INSERT INTO pregnancies (user_id, due_date, partner_name, partner_email, baby_name)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, user_id, partner_name, partner_email, due_date, conception_date, current_week, baby_name, is_active, created_at, updated_at
	`

	var pregnancy models.Pregnancy
	err := db.GetDB().QueryRow(query, userID, dueDate, partnerName, partnerEmail, babyName).Scan(
		&pregnancy.ID,
		&pregnancy.UserID,
		&pregnancy.PartnerName,
		&pregnancy.PartnerEmail,
		&pregnancy.DueDate,
		&pregnancy.ConceptionDate,
		&pregnancy.CurrentWeek,
		&pregnancy.BabyName,
		&pregnancy.IsActive,
		&pregnancy.CreatedAt,
		&pregnancy.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Create default milestones for this pregnancy
	if err := CreateDefaultMilestones(pregnancy.ID, dueDate); err != nil {
		// Log error but don't fail the pregnancy creation
		// TODO: Add proper logging
	}

	return &pregnancy, nil
}

func GetActivePregnancyByUserID(userID int) (*models.Pregnancy, error) {
	query := `
		SELECT id, user_id, partner_name, partner_email, due_date, conception_date, current_week, baby_name, is_active, created_at, updated_at
		FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`

	var pregnancy models.Pregnancy
	err := db.GetDB().QueryRow(query, userID).Scan(
		&pregnancy.ID,
		&pregnancy.UserID,
		&pregnancy.PartnerName,
		&pregnancy.PartnerEmail,
		&pregnancy.DueDate,
		&pregnancy.ConceptionDate,
		&pregnancy.CurrentWeek,
		&pregnancy.BabyName,
		&pregnancy.IsActive,
		&pregnancy.CreatedAt,
		&pregnancy.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &pregnancy, nil
}

func CreateDefaultMilestones(pregnancyID int, dueDate time.Time) error {
	// Calculate milestone dates based on due date
	firstAppointment := dueDate.AddDate(0, 0, -245) // ~35 weeks before due date
	twelveWeekScan := dueDate.AddDate(0, 0, -196)   // ~28 weeks before due date
	twentyWeekScan := dueDate.AddDate(0, 0, -140)   // ~20 weeks before due date
	thirtyFourWeekAppt := dueDate.AddDate(0, 0, -42) // ~6 weeks before due date

	milestones := []struct {
		Type          string
		Title         string
		ScheduledDate *time.Time
		WeekNumber    *int
	}{
		{models.MilestoneFirstAppointment, "First Doctor Appointment", &firstAppointment, intPtr(8)},
		{models.Milestone12WeekScan, "12 Week Scan", &twelveWeekScan, intPtr(12)},
		{models.Milestone20WeekScan, "20 Week Anatomy Scan", &twentyWeekScan, intPtr(20)},
		{models.Milestone36WeekAppointment, "36 Week Appointment", &thirtyFourWeekAppt, intPtr(36)},
		{models.MilestoneDueDate, "Due Date", &dueDate, intPtr(40)},
	}

	for _, milestone := range milestones {
		query := `
			INSERT INTO milestones (pregnancy_id, milestone_type, title, scheduled_date, week_number)
			VALUES (?, ?, ?, ?, ?)
		`
		_, err := db.GetDB().Exec(query, pregnancyID, milestone.Type, milestone.Title, milestone.ScheduledDate, milestone.WeekNumber)
		if err != nil {
			return err
		}
	}

	return nil
}

func intPtr(i int) *int {
	return &i
}

// generatePregnancyHash creates a deterministic, reversible hash for a pregnancy ID
func generatePregnancyHash(pregnancyID int) string {
	// Simple XOR-based encoding that's reversible
	const secret = 0x40202024 // System secret as hex constant (40 = @, 2024 = year)
	
	// XOR the pregnancy ID with the secret
	encoded := pregnancyID ^ secret
	
	// Convert to hex string with zero padding
	return fmt.Sprintf("%08x", encoded)
}

// validatePregnancyHash validates a hash and returns the pregnancy ID
func validatePregnancyHash(hash string) (int, error) {
	if len(hash) != 8 {
		log.Printf("Invalid hash length: %d, expected 8", len(hash))
		return 0, fmt.Errorf("invalid hash length")
	}
	
	log.Printf("Validating hash: %s", hash)
	
	// Decode the hex string
	encoded, err := strconv.ParseInt(hash, 16, 64)
	if err != nil {
		log.Printf("Invalid hex hash: %s", hash)
		return 0, fmt.Errorf("invalid hash format")
	}
	
	// Reverse the XOR to get original pregnancy ID
	const secret = 0x40202024
	pregnancyID := int(encoded) ^ secret
	
	log.Printf("Decoded pregnancy ID: %d from hash: %s", pregnancyID, hash)
	
	// Verify the pregnancy exists and is active
	pregnancy, err := GetPregnancyByID(pregnancyID)
	if err != nil {
		log.Printf("Pregnancy not found for ID: %d", pregnancyID)
		return 0, fmt.Errorf("pregnancy not found")
	}
	
	if !pregnancy.IsActive {
		log.Printf("Pregnancy %d is not active", pregnancyID)
		return 0, fmt.Errorf("pregnancy not active")
	}
	
	log.Printf("Hash validated for pregnancy ID: %d", pregnancyID)
	return pregnancyID, nil
}

func GetPregnancyByID(pregnancyID int) (*models.Pregnancy, error) {
	query := `
		SELECT id, user_id, partner_name, partner_email, due_date, conception_date, current_week, baby_name, is_active, created_at, updated_at
		FROM pregnancies 
		WHERE id = ?
		LIMIT 1
	`

	var pregnancy models.Pregnancy
	err := db.GetDB().QueryRow(query, pregnancyID).Scan(
		&pregnancy.ID,
		&pregnancy.UserID,
		&pregnancy.PartnerName,
		&pregnancy.PartnerEmail,
		&pregnancy.DueDate,
		&pregnancy.ConceptionDate,
		&pregnancy.CurrentWeek,
		&pregnancy.BabyName,
		&pregnancy.IsActive,
		&pregnancy.CreatedAt,
		&pregnancy.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &pregnancy, nil
}

func GetUserByID(userID int) (*db.User, error) {
	query := `
		SELECT id, name, email, password, is_admin, created
		FROM users 
		WHERE id = ?
		LIMIT 1
	`

	var user db.User
	err := db.GetDB().QueryRow(query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.IsAdmin,
		&user.Created,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}