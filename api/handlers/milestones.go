package handlers

import (
	"encoding/json"
	"net/http"
	"simple-go/api/db"
	"simple-go/api/middleware"
	"time"
)

// PregnancyMilestone represents a standard pregnancy milestone
type PregnancyMilestone struct {
	Week        int       `json:"week"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "appointment", "development", "milestone"
	Date        time.Time `json:"date"`
	IsPast      bool      `json:"is_past"`
	IsCurrent   bool      `json:"is_current"`
}

// GetMilestonesHandler returns pregnancy milestones for the current user
func GetMilestonesHandler(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Get current pregnancy
	var pregnancyID int
	var dueDate time.Time
	var conceptionDate *time.Time
	err := db.GetDB().QueryRow(`
		SELECT id, due_date, conception_date 
		FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE
		ORDER BY created_at DESC LIMIT 1`,
		userID).Scan(&pregnancyID, &dueDate, &conceptionDate)
	if err != nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Calculate current week
	currentWeek := calculateCurrentWeek(conceptionDate, dueDate)

	// Generate milestones based on standard pregnancy timeline
	milestones := generateMilestones(conceptionDate, dueDate, currentWeek)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(milestones)
}

func calculateCurrentWeek(conceptionDate *time.Time, dueDate time.Time) int {
	now := time.Now()
	
	if conceptionDate != nil {
		// Calculate gestational weeks from LMP (conception + 14 days)
		daysSinceConception := int(now.Sub(*conceptionDate).Hours() / 24)
		// Add 14 days to account for LMP offset
		return (daysSinceConception+14)/7 + 1
	}
	
	// Otherwise calculate from due date (40 weeks back)
	conceptionEstimate := dueDate.AddDate(0, 0, -280) // 40 weeks = 280 days
	weeks := int(now.Sub(conceptionEstimate).Hours() / 24 / 7)
	return weeks + 1
}

func generateMilestones(conceptionDate *time.Time, dueDate time.Time, currentWeek int) []PregnancyMilestone {
	// Calculate the date for a given week
	getDateForWeek := func(week int) time.Time {
		if conceptionDate != nil {
			// Calculate from conception date
			daysFromConception := (week-1)*7 - 14 // Subtract 14 days for LMP offset
			return conceptionDate.AddDate(0, 0, daysFromConception)
		}
		// Calculate from due date
		daysFromDue := (40 - week) * 7
		return dueDate.AddDate(0, 0, -daysFromDue)
	}

	// Define standard pregnancy milestones
	milestonesData := []struct {
		Week        int
		Title       string
		Description string
		Type        string
	}{
		{8, "First Prenatal Visit", "Confirm pregnancy and establish care", "appointment"},
		{10, "Baby's Heart Starts Beating", "Baby's heart begins to beat", "development"},
		{12, "12-Week Scan", "First ultrasound and genetic screening", "appointment"},
		{14, "Second Trimester Begins", "Morning sickness often improves", "milestone"},
		{16, "Gender Reveal Possible", "Sex can often be determined", "development"},
		{18, "Anatomy Scan Prep", "Prepare for detailed ultrasound", "appointment"},
		{20, "20-Week Anatomy Scan", "Detailed ultrasound examination", "appointment"},
		{20, "Halfway Point!", "You're halfway through pregnancy", "milestone"},
		{22, "Baby's Movements", "You may start feeling kicks", "development"},
		{24, "Viability Milestone", "Baby can survive outside womb", "milestone"},
		{26, "Glucose Screening", "Test for gestational diabetes", "appointment"},
		{28, "Third Trimester Begins", "Final pregnancy phase starts", "milestone"},
		{28, "28-Week Checkup", "Regular monitoring increases", "appointment"},
		{32, "32-Week Checkup", "Monitor baby's growth and position", "appointment"},
		{34, "Baby Shower Time", "Celebrate with friends and family", "milestone"},
		{36, "36-Week Checkup", "Weekly visits often begin", "appointment"},
		{37, "Full Term!", "Baby is considered full term", "milestone"},
		{38, "38-Week Checkup", "Final preparations", "appointment"},
		{39, "39-Week Checkup", "Any day now!", "appointment"},
		{40, "Due Date", "Your estimated due date", "milestone"},
		{41, "Post-Due Checkup", "Monitor if past due date", "appointment"},
	}

	milestones := make([]PregnancyMilestone, 0)
	
	for _, m := range milestonesData {
		milestone := PregnancyMilestone{
			Week:        m.Week,
			Title:       m.Title,
			Description: m.Description,
			Type:        m.Type,
			Date:        getDateForWeek(m.Week),
			IsPast:      m.Week < currentWeek,
			IsCurrent:   m.Week == currentWeek,
		}
		milestones = append(milestones, milestone)
	}

	return milestones
}