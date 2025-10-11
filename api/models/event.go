package models

import "time"

// PregnancyEvent represents an event in the pregnancy timeline
type PregnancyEvent struct {
	ID             int       `json:"id"`
	PregnancyID    int       `json:"pregnancy_id"`
	EventType      string    `json:"event_type"`
	EventTitle     string    `json:"event_title"`
	EventDescription *string  `json:"event_description"`
	EventData      *string   `json:"event_data"` // JSON string
	WeekNumber     *int      `json:"week_number"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      *int      `json:"created_by"`
}

// Event type constants for meaningful pregnancy events
const (
	EventPregnancyAnnounced   = "pregnancy_announced"    // When pregnancy first created
	EventVillagerJoined       = "villager_joined"        // Someone joins village via invite
	EventVillagerTold         = "villager_told"          // Villager marked as "knows"
	EventMilestoneReached     = "milestone_reached"      // Automatic: 12 weeks, 20 weeks, etc.
	EventAppointmentCompleted = "appointment_completed"  // Manual milestone completion
	EventUpdatePosted         = "update_posted"          // User shares news/photos
	EventWeekProgression      = "week_progression"       // Weekly automatic milestones
)

// GetEventDisplayInfo returns user-friendly display information for events
func (e *PregnancyEvent) GetEventDisplayInfo() (icon string, color string) {
	switch e.EventType {
	case EventPregnancyAnnounced:
		return "🎉", "text-pink-600"
	case EventVillagerJoined:
		return "👥", "text-blue-600"
	case EventVillagerTold:
		return "💌", "text-purple-600"
	case EventMilestoneReached:
		return "🏆", "text-yellow-600"
	case EventAppointmentCompleted:
		return "🏥", "text-green-600"
	case EventUpdatePosted:
		return "📝", "text-indigo-600"
	case EventWeekProgression:
		return "📅", "text-gray-600"
	default:
		return "📌", "text-gray-500"
	}
}
