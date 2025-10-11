package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
)

// EventService handles pregnancy event operations
type EventService struct {
	database *sql.DB
}

// NewEventService creates a new event service
func NewEventService() *EventService {
	return &EventService{
		database: db.GetDB(),
	}
}

// CreateEvent creates a new pregnancy event
func (s *EventService) CreateEvent(pregnancyID int, eventType, title, description string, weekNumber *int, createdBy *int, eventData map[string]interface{}) error {
	var eventDataJSON *string
	if eventData != nil {
		data, err := json.Marshal(eventData)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}
		dataStr := string(data)
		eventDataJSON = &dataStr
	}

	query := `
		INSERT INTO pregnancy_events (pregnancy_id, event_type, event_title, event_description, event_data, week_number, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.database.Exec(query, pregnancyID, eventType, title, description, eventDataJSON, weekNumber, createdBy)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	log.Printf("Created event: %s for pregnancy %d", eventType, pregnancyID)
	return nil
}

// GetTimelineEvents retrieves events for a pregnancy timeline
func (s *EventService) GetTimelineEvents(pregnancyID int, limit, offset int) ([]*models.PregnancyEvent, error) {
	query := `
		SELECT id, pregnancy_id, event_type, event_title, event_description, event_data, week_number, created_at, created_by
		FROM pregnancy_events 
		WHERE pregnancy_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`

	rows, err := s.database.Query(query, pregnancyID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*models.PregnancyEvent
	for rows.Next() {
		event := &models.PregnancyEvent{}
		err := rows.Scan(
			&event.ID,
			&event.PregnancyID,
			&event.EventType,
			&event.EventTitle,
			&event.EventDescription,
			&event.EventData,
			&event.WeekNumber,
			&event.CreatedAt,
			&event.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// GetTimelineEventsHandler returns timeline events for the current user's pregnancy
func GetTimelineEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's active pregnancy (either as owner or partner)
	pregnancy, err := GetActivePregnancyForUser(claims.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if pregnancy == nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Parse limit and offset from query params
	limit := 20 // default
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0 // default
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get timeline events
	eventService := NewEventService()
	events, err := eventService.GetTimelineEvents(pregnancy.ID, limit, offset)
	if err != nil {
		log.Printf("Failed to get timeline events: %v", err)
		http.Error(w, "Failed to retrieve timeline events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": events,
		"total":  len(events),
	})
}

// Helper functions for creating specific event types

// CreatePregnancyAnnouncedEvent creates the initial pregnancy announcement event
func CreatePregnancyAnnouncedEvent(pregnancyID int, userID int, weekNumber *int) error {
	eventService := NewEventService()
	return eventService.CreateEvent(
		pregnancyID,
		models.EventPregnancyAnnounced,
		"Pregnancy begins! 🎉",
		"Your pregnancy tracking has been set up and your journey begins!",
		weekNumber,
		&userID,
		nil,
	)
}

// CreateVillagerJoinedEvent creates an event when someone joins the village via invite
func CreateVillagerJoinedEvent(pregnancyID int, villagerName, relationship string, weekNumber *int) error {
	eventService := NewEventService()
	
	eventData := map[string]interface{}{
		"villager_name": villagerName,
		"relationship":  relationship,
	}

	return eventService.CreateEvent(
		pregnancyID,
		models.EventVillagerJoined,
		fmt.Sprintf("%s joined your village", villagerName),
		fmt.Sprintf("%s (%s) has joined your pregnancy village through your invite link", villagerName, relationship),
		weekNumber,
		nil, // No specific user created this (it's from invite)
		eventData,
	)
}

// CreateVillagerAddedEvent creates an event when someone is manually added to the village
func CreateVillagerAddedEvent(pregnancyID int, villagerName, relationship string, weekNumber *int) error {
	eventService := NewEventService()
	
	eventData := map[string]interface{}{
		"villager_name": villagerName,
		"relationship":  relationship,
	}

	return eventService.CreateEvent(
		pregnancyID,
		models.EventVillagerJoined, // Use same event type as joined
		fmt.Sprintf("Added %s to your village", villagerName),
		fmt.Sprintf("%s (%s) has been added to your pregnancy village", villagerName, relationship),
		weekNumber,
		nil, // System generated
		eventData,
	)
}

// CreateVillagerToldEvent creates an event when a villager is marked as knowing
func CreateVillagerToldEvent(pregnancyID int, villagerName string, userID int, weekNumber *int) error {
	eventService := NewEventService()
	
	eventData := map[string]interface{}{
		"villager_name": villagerName,
	}

	return eventService.CreateEvent(
		pregnancyID,
		models.EventVillagerTold,
		fmt.Sprintf("Told %s about pregnancy", villagerName),
		fmt.Sprintf("%s now knows about your pregnancy", villagerName),
		weekNumber,
		&userID,
		eventData,
	)
}

// CreateMilestoneReachedEvent creates an event for reaching pregnancy milestones
func CreateMilestoneReachedEvent(pregnancyID int, milestoneTitle string, weekNumber int) error {
	eventService := NewEventService()
	
	eventData := map[string]interface{}{
		"milestone_week": weekNumber,
	}

	return eventService.CreateEvent(
		pregnancyID,
		models.EventMilestoneReached,
		fmt.Sprintf("Week %d milestone: %s", weekNumber, milestoneTitle),
		fmt.Sprintf("You've reached week %d of your pregnancy!", weekNumber),
		&weekNumber,
		nil, // System generated
		eventData,
	)
}

// CreateUpdateSharedEvent creates an event when a pregnancy update is shared
func CreateUpdateSharedEvent(pregnancyID int, userID int, updateTitle, updateDescription string, weekNumber *int) error {
	eventService := NewEventService()
	
	eventData := map[string]interface{}{
		"update_title": updateTitle,
	}

	return eventService.CreateEvent(
		pregnancyID,
		models.EventUpdatePosted,
		updateTitle,
		updateDescription,
		weekNumber,
		&userID,
		eventData,
	)
}
