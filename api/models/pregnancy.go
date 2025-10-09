package models

import (
	"time"
)

type Pregnancy struct {
	ID                 int       `json:"id" db:"id"`
	UserID             int       `json:"user_id" db:"user_id"`
	PartnerName        *string   `json:"partner_name" db:"partner_name"`
	PartnerEmail       *string   `json:"partner_email" db:"partner_email"`
	DueDate            time.Time `json:"due_date" db:"due_date"`
	ConceptionDate     *time.Time `json:"conception_date" db:"conception_date"`
	CurrentWeek        int       `json:"current_week" db:"current_week"`
	BabyName           *string   `json:"baby_name" db:"baby_name"`
	IsActive           bool      `json:"is_active" db:"is_active"`
	ShareID            string    `json:"share_id" db:"share_id"`
	CoverPhotoFilename *string   `json:"cover_photo_filename" db:"cover_photo_filename"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// PregnancyWithUser includes user information
type PregnancyWithUser struct {
	Pregnancy
	Username string `json:"username" db:"username"`
	Email    string `json:"email" db:"email"`
}

// GetCurrentWeek calculates the current pregnancy week based on conception date or due date
func (p *Pregnancy) GetCurrentWeek() int {
	now := time.Now()
	
	// If we have conception date, calculate gestational weeks from LMP (conception + 14 days)
	if p.ConceptionDate != nil {
		daysSinceConception := int(now.Sub(*p.ConceptionDate).Hours() / 24)
		// Add 14 days to account for LMP offset (gestational age calculation)
		weeks := (daysSinceConception + 14) / 7
		return weeks + 1 // Pregnancy weeks start at 1
	}
	
	// Otherwise calculate from due date (40 weeks back from LMP perspective)
	// Due date is calculated from LMP, so we can use it directly
	conceptionEstimate := p.DueDate.AddDate(0, 0, -280) // 40 weeks = 280 days
	weeks := int(now.Sub(conceptionEstimate).Hours() / 24 / 7)
	return weeks + 1
}

// GetWeeksRemaining calculates weeks remaining until due date
func (p *Pregnancy) GetWeeksRemaining() int {
	now := time.Now()
	if p.DueDate.Before(now) {
		return 0 // Baby is due or born
	}
	
	daysRemaining := int(p.DueDate.Sub(now).Hours() / 24)
	return (daysRemaining + 6) / 7 // Round up to nearest week
}

// IsOverdue checks if pregnancy is past due date
func (p *Pregnancy) IsOverdue() bool {
	return time.Now().After(p.DueDate)
}