package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"simple-go/api/db"
	"simple-go/api/models"
)

// PublicTimelineItem represents a public timeline item for village members
type PublicTimelineItem struct {
	ID          int                  `json:"id"`
	Title       string               `json:"title"`
	Description *string              `json:"description"`
	WeekNumber  *int                 `json:"week_number"`
	UpdateDate  string               `json:"update_date"`
	CreatedBy   string               `json:"created_by"`
	Photos      []models.UpdatePhoto `json:"photos,omitempty"`
	PregnancyID int                  `json:"pregnancy_id"`
}

// PublicTimelineHandler returns shared updates for a pregnancy via share ID
func PublicTimelineHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract share ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/timeline/")
	shareID := path
	
	if shareID == "" {
		http.Error(w, "Invalid share ID", http.StatusBadRequest)
		return
	}

	// Get pregnancy by share ID
	pregnancy, err := GetPregnancyByShareID(shareID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Pregnancy not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
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

	// Get public timeline items (only shared updates)
	items, err := getPublicTimelineItems(pregnancy.ID, limit, offset)
	if err != nil {
		log.Printf("Failed to get public timeline items: %v", err)
		http.Error(w, "Failed to retrieve timeline", http.StatusInternalServerError)
		return
	}

	// Get pregnancy info for context
	var userName string
	err = db.GetDB().QueryRow(`SELECT name FROM users WHERE id = ?`, pregnancy.UserID).Scan(&userName)
	if err != nil {
		log.Printf("Failed to get user name: %v", err)
		userName = "Parent"
	}

	parentNames := userName
	if pregnancy.PartnerName != nil && *pregnancy.PartnerName != "" {
		parentNames = parentNames + " & " + *pregnancy.PartnerName
	}

	babyName := "Baby"
	if pregnancy.BabyName != nil && *pregnancy.BabyName != "" {
		babyName = *pregnancy.BabyName
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"updates":      items,
		"total":        len(items),
		"pregnancy": map[string]interface{}{
			"parent_names":  parentNames,
			"baby_name":     babyName,
			"due_date":      pregnancy.DueDate.Format("2006-01-02"),
			"current_week":  pregnancy.GetCurrentWeek(),
		},
	})
}

// getPublicTimelineItems fetches only shared updates for public viewing
func getPublicTimelineItems(pregnancyID, limit, offset int) ([]PublicTimelineItem, error) {
	query := `
	SELECT 
		pu.id,
		pu.title,
		pu.content as description,
		pu.week_number,
		COALESCE(pu.update_date, pu.created_at) as update_date,
		u.name as created_by
	FROM pregnancy_updates pu
	JOIN pregnancies p ON p.id = pu.pregnancy_id
	JOIN users u ON u.id = p.user_id
	WHERE pu.pregnancy_id = ? AND pu.is_shared = TRUE
	ORDER BY update_date DESC, pu.id DESC
	LIMIT ? OFFSET ?`

	rows, err := db.GetDB().Query(query, pregnancyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PublicTimelineItem
	for rows.Next() {
		var item PublicTimelineItem
		var createdBy string
		var updateDateStr string
		
		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.WeekNumber,
			&updateDateStr,
			&createdBy,
		)
		if err != nil {
			continue
		}

		// Parse the datetime from SQLite and format as ISO string
		if updateDateStr != "" {
			// SQLite datetime format is typically "2006-01-02 15:04:05" or "2006-01-02T15:04:05Z"
			var parsedTime time.Time
			
			// Try common SQLite datetime formats
			formats := []string{
				time.RFC3339,
				"2006-01-02 15:04:05",
				"2006-01-02T15:04:05",
				"2006-01-02",
			}
			
			for _, format := range formats {
				if parsed, err := time.Parse(format, updateDateStr); err == nil {
					parsedTime = parsed
					break
				}
			}
			
			// If we couldn't parse it, use the original string
			if !parsedTime.IsZero() {
				// Ensure it's UTC and format as ISO string
				item.UpdateDate = parsedTime.UTC().Format(time.RFC3339)
			} else {
				item.UpdateDate = updateDateStr
			}
		}

		item.CreatedBy = createdBy
		item.PregnancyID = pregnancyID

		// Get photos/videos for this update
		photos, _ := getUpdatePhotos(item.ID)
		item.Photos = photos

		items = append(items, item)
	}

	return items, nil
}