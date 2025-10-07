package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
)

type CreatePregnancyRequest struct {
	DueDate        string  `json:"due_date"`
	PartnerName    *string `json:"partner_name"`
	BabyName       *string `json:"baby_name"`
	ConceptionDate *string `json:"conception_date"`
}

type PregnancyResponse struct {
	*models.Pregnancy
	CurrentWeek int `json:"current_week_calculated"`
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

	// Parse conception date if provided
	var conceptionDate *time.Time
	if req.ConceptionDate != nil && *req.ConceptionDate != "" {
		cd, err := time.Parse("2006-01-02", *req.ConceptionDate)
		if err != nil {
			http.Error(w, "Invalid conception date format", http.StatusBadRequest)
			return
		}
		conceptionDate = &cd
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
	pregnancy, err := CreatePregnancy(claims.UserID, dueDate, conceptionDate, req.PartnerName, req.BabyName)
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

	response := &PregnancyResponse{
		Pregnancy:   pregnancy,
		CurrentWeek: pregnancy.GetCurrentWeek(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Database functions

func CreatePregnancy(userID int, dueDate time.Time, conceptionDate *time.Time, partnerName, babyName *string) (*models.Pregnancy, error) {
	query := `
		INSERT INTO pregnancies (user_id, due_date, conception_date, partner_name, baby_name)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, user_id, partner_name, due_date, conception_date, current_week, baby_name, is_active, created_at, updated_at
	`

	var pregnancy models.Pregnancy
	err := db.GetDB().QueryRow(query, userID, dueDate, conceptionDate, partnerName, babyName).Scan(
		&pregnancy.ID,
		&pregnancy.UserID,
		&pregnancy.PartnerName,
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
		SELECT id, user_id, partner_name, due_date, conception_date, current_week, baby_name, is_active, created_at, updated_at
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