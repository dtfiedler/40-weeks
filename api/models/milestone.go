package models

import (
	"fmt"
	"time"
)

type Milestone struct {
	ID             int       `json:"id" db:"id"`
	PregnancyID    int       `json:"pregnancy_id" db:"pregnancy_id"`
	MilestoneType  string    `json:"milestone_type" db:"milestone_type"`
	Title          string    `json:"title" db:"title"`
	ScheduledDate  *time.Time `json:"scheduled_date" db:"scheduled_date"`
	CompletedDate  *time.Time `json:"completed_date" db:"completed_date"`
	IsCompleted    bool      `json:"is_completed" db:"is_completed"`
	Notes          *string   `json:"notes" db:"notes"`
	WeekNumber     *int      `json:"week_number" db:"week_number"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// Milestone types
const (
	MilestoneFirstAppointment = "first_appointment"
	Milestone12WeekScan       = "12_week_scan"
	Milestone20WeekScan       = "20_week_scan"
	Milestone36WeekAppointment = "36_week_appointment"
	MilestoneDueDate          = "due_date"
	MilestoneInductionScheduled = "induction_scheduled"
	MilestoneAnnouncementMade = "announcement_made"
	MilestoneGenderRevealed   = "gender_revealed"
	MilestoneNurseryComplete  = "nursery_complete"
	MilestoneHospitalBagPacked = "hospital_bag_packed"
	MilestoneMaternityLeave   = "maternity_leave"
	MilestonePaternityLeave   = "paternity_leave"
	MilestoneBabyShower       = "baby_shower"
)

// GetDisplayTitle returns a formatted milestone title
func (m *Milestone) GetDisplayTitle() string {
	switch m.MilestoneType {
	case MilestoneFirstAppointment:
		return "First Doctor Appointment"
	case Milestone12WeekScan:
		return "12 Week Scan"
	case Milestone20WeekScan:
		return "20 Week Anatomy Scan"
	case Milestone36WeekAppointment:
		return "36 Week Appointment"
	case MilestoneDueDate:
		return "Due Date"
	case MilestoneInductionScheduled:
		return "Induction Scheduled"
	case MilestoneAnnouncementMade:
		return "Pregnancy Announcement"
	case MilestoneGenderRevealed:
		return "Gender Reveal"
	case MilestoneNurseryComplete:
		return "Nursery Complete"
	case MilestoneHospitalBagPacked:
		return "Hospital Bag Packed"
	case MilestoneMaternityLeave:
		return "Maternity Leave Starts"
	case MilestonePaternityLeave:
		return "Paternity Leave Starts"
	case MilestoneBabyShower:
		return "Baby Shower"
	default:
		return m.Title
	}
}

// IsScheduled checks if the milestone has a scheduled date
func (m *Milestone) IsScheduled() bool {
	return m.ScheduledDate != nil
}

// IsOverdue checks if the milestone is past its scheduled date and not completed
func (m *Milestone) IsOverdue() bool {
	if m.IsCompleted || m.ScheduledDate == nil {
		return false
	}
	return time.Now().After(*m.ScheduledDate)
}

// DaysUntilScheduled returns the number of days until the scheduled date
func (m *Milestone) DaysUntilScheduled() int {
	if m.ScheduledDate == nil {
		return -1
	}
	
	days := int(m.ScheduledDate.Sub(time.Now()).Hours() / 24)
	return days
}

// GetStatusText returns a human-readable status for the milestone
func (m *Milestone) GetStatusText() string {
	if m.IsCompleted {
		return "Completed"
	}
	
	if m.ScheduledDate == nil {
		return "Not Scheduled"
	}
	
	if m.IsOverdue() {
		return "Overdue"
	}
	
	days := m.DaysUntilScheduled()
	if days == 0 {
		return "Today"
	} else if days == 1 {
		return "Tomorrow"
	} else if days > 0 {
		return fmt.Sprintf("In %d days", days)
	} else {
		return "Past Due"
	}
}
