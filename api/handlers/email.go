package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"simple-go/api/db"
	"simple-go/api/middleware"
	"simple-go/api/models"
	"simple-go/api/services/email"
	"strconv"
	"time"
)

// SendTestEmailRequest represents a test email request
type SendTestEmailRequest struct {
	ToEmail string `json:"to_email"`
	ToName  string `json:"to_name"`
}

// SendTestEmailHandler sends a test email to verify configuration
func SendTestEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SendTestEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ToEmail == "" {
		http.Error(w, "to_email is required", http.StatusBadRequest)
		return
	}

	if req.ToName == "" {
		req.ToName = "Test User"
	}

	// Create email service
	emailService, err := email.NewEmailService()
	if err != nil {
		http.Error(w, "Failed to initialize email service", http.StatusInternalServerError)
		return
	}

	// Send test email
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = emailService.SendTestEmail(ctx, req.ToEmail, req.ToName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send test email: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Test email sent successfully to %s", req.ToEmail),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEmailNotificationsHandler returns email notification history
func GetEmailNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from auth context
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's pregnancy
	var pregnancyID int
	err := db.GetDB().QueryRow(`
		SELECT id FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE 
		ORDER BY created_at DESC LIMIT 1`,
		claims.UserID,
	).Scan(&pregnancyID)
	
	if err != nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Parse query parameters
	limit := 50
	offset := 0
	
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get email notifications
	query := `
		SELECT 
			en.id, en.pregnancy_id, en.village_member_id, en.update_id, en.milestone_id,
			en.email_type, en.subject, en.sent_at, en.delivery_status, en.ses_message_id,
			en.created_at, vm.name as recipient_name, vm.email as recipient_email
		FROM email_notifications en
		JOIN village_members vm ON en.village_member_id = vm.id
		WHERE en.pregnancy_id = ?
		ORDER BY en.sent_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := db.GetDB().Query(query, pregnancyID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to fetch notifications", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var notifications []map[string]interface{}
	for rows.Next() {
		var notification models.EmailNotification
		var recipientName, recipientEmail string
		
		err := rows.Scan(
			&notification.ID,
			&notification.PregnancyID,
			&notification.VillageMemberID,
			&notification.UpdateID,
			&notification.MilestoneID,
			&notification.EmailType,
			&notification.Subject,
			&notification.SentAt,
			&notification.DeliveryStatus,
			&notification.SESMessageID,
			&notification.CreatedAt,
			&recipientName,
			&recipientEmail,
		)
		
		if err != nil {
			continue
		}

		notificationData := map[string]interface{}{
			"id":               notification.ID,
			"email_type":       notification.EmailType,
			"display_type":     notification.GetDisplayEmailType(),
			"subject":          notification.Subject,
			"recipient_name":   recipientName,
			"recipient_email":  recipientEmail,
			"sent_at":          notification.SentAt,
			"delivery_status":  notification.DeliveryStatus,
			"display_status":   notification.GetDisplayDeliveryStatus(),
			"is_successful":    notification.IsSuccessful(),
			"is_failed":        notification.IsFailed(),
			"ses_message_id":   notification.SESMessageID,
			"created_at":       notification.CreatedAt,
		}
		
		notifications = append(notifications, notificationData)
	}

	response := map[string]interface{}{
		"notifications": notifications,
		"total":         len(notifications),
		"limit":         limit,
		"offset":        offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetEmailStatisticsHandler returns email sending statistics
func GetEmailStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from auth context
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user's pregnancy
	var pregnancyID int
	err := db.GetDB().QueryRow(`
		SELECT id FROM pregnancies 
		WHERE user_id = ? AND is_active = TRUE 
		ORDER BY created_at DESC LIMIT 1`,
		claims.UserID,
	).Scan(&pregnancyID)
	
	if err != nil {
		http.Error(w, "No active pregnancy found", http.StatusNotFound)
		return
	}

	// Create email service
	emailService, err := email.NewEmailService()
	if err != nil {
		http.Error(w, "Failed to initialize email service", http.StatusInternalServerError)
		return
	}

	// Get statistics
	stats, err := emailService.GetEmailStatistics(pregnancyID)
	if err != nil {
		http.Error(w, "Failed to get email statistics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// TestEmailConfigurationHandler tests the email service configuration
func TestEmailConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create email service
	emailService, err := email.NewEmailService()
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Failed to initialize email service: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Test configuration
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = emailService.TestEmailConfiguration(ctx)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Email configuration test failed: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Email configuration is working correctly",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SendUpdateNotificationHandler manually triggers update notification emails
func SendUpdateNotificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from auth context
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req struct {
		UpdateID int `json:"update_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UpdateID == 0 {
		http.Error(w, "update_id is required", http.StatusBadRequest)
		return
	}

	// Get the update and verify ownership
	var update models.PregnancyUpdate
	var pregnancy models.Pregnancy
	
	query := `
		SELECT 
			pu.id, pu.pregnancy_id, pu.title, pu.content, pu.week_number, 
			pu.update_date, pu.is_shared, pu.created_at,
			p.id, p.user_id, p.due_date, p.conception_date, p.baby_name, 
			p.partner_name, p.partner_email, p.share_id, p.is_active, p.created_at
		FROM pregnancy_updates pu
		JOIN pregnancies p ON pu.pregnancy_id = p.id
		WHERE pu.id = ? AND p.user_id = ?
	`
	
	err := db.GetDB().QueryRow(query, req.UpdateID, claims.UserID).Scan(
		&update.ID,
		&update.PregnancyID,
		&update.Title,
		&update.Content,
		&update.WeekNumber,
		&update.UpdateDate,
		&update.IsShared,
		&update.CreatedAt,
		&pregnancy.ID,
		&pregnancy.UserID,
		&pregnancy.DueDate,
		&pregnancy.ConceptionDate,
		&pregnancy.BabyName,
		&pregnancy.PartnerName,
		&pregnancy.PartnerEmail,
		&pregnancy.ShareID,
		&pregnancy.IsActive,
		&pregnancy.CreatedAt,
	)
	
	if err != nil {
		http.Error(w, "Update not found or access denied", http.StatusNotFound)
		return
	}

	// Create email service
	emailService, err := email.NewEmailService()
	if err != nil {
		http.Error(w, "Failed to initialize email service", http.StatusInternalServerError)
		return
	}

	// Send notification
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	err = emailService.SendUpdateNotification(ctx, &update, &pregnancy)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send notification: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Update notification sent successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}