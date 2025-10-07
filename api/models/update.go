package models

import (
	"fmt"
	"time"
)

type PregnancyUpdate struct {
	ID              int       `json:"id" db:"id"`
	PregnancyID     int       `json:"pregnancy_id" db:"pregnancy_id"`
	WeekNumber      *int      `json:"week_number" db:"week_number"`
	Title           string    `json:"title" db:"title"`
	Content         *string   `json:"content" db:"content"`
	UpdateType      string    `json:"update_type" db:"update_type"`
	AppointmentType *string   `json:"appointment_type" db:"appointment_type"`
	IsShared        bool      `json:"is_shared" db:"is_shared"`
	SharedAt        *time.Time `json:"shared_at" db:"shared_at"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	Photos          []UpdatePhoto `json:"photos,omitempty"`
}

type UpdatePhoto struct {
	ID               int       `json:"id" db:"id"`
	UpdateID         int       `json:"update_id" db:"update_id"`
	Filename         string    `json:"filename" db:"filename"`
	OriginalFilename string    `json:"original_filename" db:"original_filename"`
	FileSize         *int      `json:"file_size" db:"file_size"`
	Caption          *string   `json:"caption" db:"caption"`
	SortOrder        int       `json:"sort_order" db:"sort_order"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Update types
const (
	UpdateTypeGeneral     = "general"
	UpdateTypeAppointment = "appointment"
	UpdateTypeMilestone   = "milestone"
	UpdateTypePhoto       = "photo"
)

// Appointment types
const (
	AppointmentTypeFirst      = "first_appointment"
	AppointmentType12Week     = "12_week_scan"
	AppointmentType20Week     = "20_week_scan"
	AppointmentType28Week     = "28_week_checkup"
	AppointmentType32Week     = "32_week_checkup"
	AppointmentType36Week     = "36_week_checkup"
	AppointmentType38Week     = "38_week_checkup"
	AppointmentType40Week     = "40_week_checkup"
	AppointmentTypeUltrasound = "ultrasound"
	AppointmentTypeBloodwork  = "bloodwork"
	AppointmentTypeOther      = "other"
)

// GetDisplayUpdateType returns a formatted update type string
func (pu *PregnancyUpdate) GetDisplayUpdateType() string {
	switch pu.UpdateType {
	case UpdateTypeGeneral:
		return "General Update"
	case UpdateTypeAppointment:
		return "Doctor Appointment"
	case UpdateTypeMilestone:
		return "Milestone"
	case UpdateTypePhoto:
		return "Photo Update"
	default:
		return "Update"
	}
}

// GetDisplayAppointmentType returns a formatted appointment type string
func (pu *PregnancyUpdate) GetDisplayAppointmentType() string {
	if pu.AppointmentType == nil {
		return ""
	}
	
	switch *pu.AppointmentType {
	case AppointmentTypeFirst:
		return "First Appointment"
	case AppointmentType12Week:
		return "12 Week Scan"
	case AppointmentType20Week:
		return "20 Week Anatomy Scan"
	case AppointmentType28Week:
		return "28 Week Checkup"
	case AppointmentType32Week:
		return "32 Week Checkup"
	case AppointmentType36Week:
		return "36 Week Checkup"
	case AppointmentType38Week:
		return "38 Week Checkup"
	case AppointmentType40Week:
		return "40 Week Checkup"
	case AppointmentTypeUltrasound:
		return "Ultrasound"
	case AppointmentTypeBloodwork:
		return "Bloodwork"
	default:
		return "Doctor Appointment"
	}
}

// GetWeekDisplay returns a formatted week number string
func (pu *PregnancyUpdate) GetWeekDisplay() string {
	if pu.WeekNumber == nil {
		return ""
	}
	
	if *pu.WeekNumber == 1 {
		return "Week 1"
	}
	
	return fmt.Sprintf("Week %d", *pu.WeekNumber)
}

// HasPhotos checks if the update has any photos
func (pu *PregnancyUpdate) HasPhotos() bool {
	return len(pu.Photos) > 0
}

// GetPhotoCount returns the number of photos in the update
func (pu *PregnancyUpdate) GetPhotoCount() int {
	return len(pu.Photos)
}