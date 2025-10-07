package models

import (
	"time"
)

type EmailNotification struct {
	ID               int       `json:"id" db:"id"`
	PregnancyID      int       `json:"pregnancy_id" db:"pregnancy_id"`
	VillageMemberID  int       `json:"village_member_id" db:"village_member_id"`
	UpdateID         *int      `json:"update_id" db:"update_id"`
	MilestoneID      *int      `json:"milestone_id" db:"milestone_id"`
	EmailType        string    `json:"email_type" db:"email_type"`
	Subject          string    `json:"subject" db:"subject"`
	SentAt           time.Time `json:"sent_at" db:"sent_at"`
	DeliveryStatus   string    `json:"delivery_status" db:"delivery_status"`
	SESMessageID     *string   `json:"ses_message_id" db:"ses_message_id"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Email types
const (
	EmailTypeUpdate       = "update"
	EmailTypeMilestone    = "milestone"
	EmailTypeAnnouncement = "announcement"
	EmailTypeWelcome      = "welcome"
	EmailTypeReminder     = "reminder"
)

// Delivery statuses
const (
	DeliveryStatusSent      = "sent"
	DeliveryStatusDelivered = "delivered"
	DeliveryStatusBounced   = "bounced"
	DeliveryStatusFailed    = "failed"
	DeliveryStatusComplaint = "complaint"
)

// GetDisplayEmailType returns a formatted email type string
func (en *EmailNotification) GetDisplayEmailType() string {
	switch en.EmailType {
	case EmailTypeUpdate:
		return "Pregnancy Update"
	case EmailTypeMilestone:
		return "Milestone"
	case EmailTypeAnnouncement:
		return "Announcement"
	case EmailTypeWelcome:
		return "Welcome"
	case EmailTypeReminder:
		return "Reminder"
	default:
		return "Email"
	}
}

// GetDisplayDeliveryStatus returns a formatted delivery status string
func (en *EmailNotification) GetDisplayDeliveryStatus() string {
	switch en.DeliveryStatus {
	case DeliveryStatusSent:
		return "Sent"
	case DeliveryStatusDelivered:
		return "Delivered"
	case DeliveryStatusBounced:
		return "Bounced"
	case DeliveryStatusFailed:
		return "Failed"
	case DeliveryStatusComplaint:
		return "Complaint"
	default:
		return "Unknown"
	}
}

// IsSuccessful checks if the email was successfully delivered
func (en *EmailNotification) IsSuccessful() bool {
	return en.DeliveryStatus == DeliveryStatusSent || en.DeliveryStatus == DeliveryStatusDelivered
}

// IsFailed checks if the email delivery failed
func (en *EmailNotification) IsFailed() bool {
	return en.DeliveryStatus == DeliveryStatusBounced || 
		   en.DeliveryStatus == DeliveryStatusFailed ||
		   en.DeliveryStatus == DeliveryStatusComplaint
}

// NotificationSummary provides summary statistics for email notifications
type NotificationSummary struct {
	TotalSent      int `json:"total_sent"`
	TotalDelivered int `json:"total_delivered"`
	TotalFailed    int `json:"total_failed"`
	TotalBounced   int `json:"total_bounced"`
	DeliveryRate   float64 `json:"delivery_rate"`
}