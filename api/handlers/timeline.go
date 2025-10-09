package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
)

// TimelineItem represents a combined timeline item (event or update)
type TimelineItem struct {
	ID          int                      `json:"id"`
	Type        string                   `json:"type"` // "event" or "update"
	Title       string                   `json:"title"`
	Description *string                  `json:"description"`
	WeekNumber  *int                     `json:"week_number"`
	CreatedAt   string                   `json:"created_at"`
	EventType   *string                  `json:"event_type,omitempty"`
	UpdateType  *string                  `json:"update_type,omitempty"`
	Photos      []models.UpdatePhoto     `json:"photos,omitempty"`
	IsShared    *bool                    `json:"is_shared,omitempty"`
	PregnancyID int                      `json:"pregnancy_id"`
	CreatedBy   *string                  `json:"created_by,omitempty"` // User name who created this item
}

// GetCombinedTimelineHandler returns a combined timeline of events and updates
func GetCombinedTimelineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's active pregnancy
	pregnancy, err := GetActivePregnancyForUser(claims.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if pregnancy == nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Parse limit and offset
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get combined timeline items
	items, err := getCombinedTimelineItems(pregnancy.ID, limit, offset)
	if err != nil {
		log.Printf("Failed to get timeline items: %v", err)
		http.Error(w, "Failed to retrieve timeline", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events": items,
		"total":  len(items),
	})
}

// getCombinedTimelineItems fetches and combines events and updates into a single timeline
func getCombinedTimelineItems(pregnancyID, limit, offset int) ([]TimelineItem, error) {
	// Query to get both events and updates, but exclude update_posted events since we show the actual updates
	query := `
	SELECT 
		'event' as type,
		pe.id as item_id,
		pe.event_title as title,
		pe.event_description as description,
		pe.event_type,
		NULL as update_type,
		NULL as is_shared,
		pe.week_number,
		pe.created_at as sort_date,
		pe.created_at,
		u.name as created_by
	FROM pregnancy_events pe
	LEFT JOIN users u ON pe.created_by = u.id
	WHERE pe.pregnancy_id = ? AND pe.event_type != 'update_posted'
	
	UNION ALL
	
	SELECT 
		'update' as type,
		pu.id as item_id,
		pu.title,
		pu.content as description,
		NULL as event_type,
		pu.update_type,
		pu.is_shared,
		pu.week_number,
		COALESCE(pu.update_date, pu.created_at) as sort_date,
		COALESCE(pu.update_date, pu.created_at) as created_at,
		u.name as created_by
	FROM pregnancy_updates pu
	JOIN pregnancies p ON p.id = pu.pregnancy_id
	JOIN users u ON u.id = p.user_id
	WHERE pu.pregnancy_id = ?
	
	ORDER BY sort_date DESC, item_id DESC
	LIMIT ? OFFSET ?`

	rows, err := db.GetDB().Query(query, pregnancyID, pregnancyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TimelineItem
	for rows.Next() {
		var item TimelineItem
		var eventType, updateType *string
		var isShared *bool
		var createdBy *string
		var sortDate string  // We don't need to store this, just scan it
		
		var itemID int
		err := rows.Scan(
			&item.Type,
			&itemID,
			&item.Title,
			&item.Description,
			&eventType,
			&updateType,
			&isShared,
			&item.WeekNumber,
			&sortDate,        // Sort date (we scan but don't use)
			&item.CreatedAt,  // Actual created_at for display
			&createdBy,
		)
		if err != nil {
			continue
		}
		
		item.ID = itemID

		item.EventType = eventType
		item.UpdateType = updateType
		item.IsShared = isShared
		item.PregnancyID = pregnancyID
		item.CreatedBy = createdBy

		// If this is an update, get its photos
		if item.Type == "update" {
			photos, _ := getUpdatePhotos(item.ID)
			item.Photos = photos
		}

		items = append(items, item)
	}

	return items, nil
}

// getUpdatePhotos fetches photos for a specific update
func getUpdatePhotos(updateID int) ([]models.UpdatePhoto, error) {
	rows, err := db.GetDB().Query(`
		SELECT id, update_id, filename, original_filename, file_size, caption, sort_order, created_at
		FROM update_photos 
		WHERE update_id = ? 
		ORDER BY sort_order`, updateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []models.UpdatePhoto
	for rows.Next() {
		var photo models.UpdatePhoto
		err := rows.Scan(&photo.ID, &photo.UpdateID, &photo.Filename, &photo.OriginalFilename,
			&photo.FileSize, &photo.Caption, &photo.SortOrder, &photo.CreatedAt)
		if err != nil {
			continue
		}
		photos = append(photos, photo)
	}

	return photos, nil
}