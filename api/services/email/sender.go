package email

import (
	"context"
	"fmt"
	"log"
	"simple-go/api/db"
	"simple-go/api/models"
	"strings"
	"time"
)

// SendUpdateNotification sends email notifications for pregnancy updates
func (e *EmailService) SendUpdateNotification(ctx context.Context, update *models.PregnancyUpdate, pregnancy *models.Pregnancy) error {
	if !e.config.EmailEnabled {
		log.Printf("Email disabled, skipping update notification for pregnancy %d", pregnancy.ID)
		return nil
	}

	// Get village members for this pregnancy
	villageMembers, err := e.getVillageMembers(pregnancy.ID)
	if err != nil {
		return fmt.Errorf("failed to get village members: %w", err)
	}

	if len(villageMembers) == 0 {
		log.Printf("No village members found for pregnancy %d", pregnancy.ID)
		return nil
	}

	// Get update photos
	photos := e.getUpdatePhotos(update.ID)
	photoCount := len(photos)

	// Generate timeline URL
	timelineURL := fmt.Sprintf("%s/view/%s", e.getBaseURL(), pregnancy.ShareID)
	
	// Generate first photo URL if available
	firstPhotoURL := ""
	if photoCount > 0 {
		firstPhotoURL = fmt.Sprintf("%s/images/%d/%s", e.getBaseURL(), pregnancy.ID, photos[0].Filename)
	}

	// Prepare template data
	templateData := &TemplateData{
		SenderName:      e.config.SenderName,
		PregnancyID:     pregnancy.ID,
		ParentNames:     e.getParentNames(pregnancy),
		DueDate:         pregnancy.DueDate.Format("January 2, 2006"),
		CurrentWeek:     pregnancy.GetCurrentWeek(),
		TimelineURL:     timelineURL,
		Update:          update,
		UpdateTitle:     update.Title,
		UpdateContent:   getStringValue(update.Content),
		UpdateWeek:      *update.WeekNumber,
		UpdateDate:      update.UpdateDate.Format("January 2, 2006"),
		UpdatePhotos:    make([]string, photoCount), // Just for count
		FirstPhotoURL:   firstPhotoURL,
	}

	// Send emails to all village members
	for _, member := range villageMembers {
		// Use only the first name for a more personal touch
		firstName := strings.Fields(strings.TrimSpace(member.Name))[0]
		templateData.RecipientName = firstName
		
		// Generate email content for this recipient
		htmlContent, textContent, err := e.UpdateNotificationTemplate(templateData)
		if err != nil {
			log.Printf("Failed to generate email template for %s: %v", member.Email, err)
			continue
		}
		
		subject := e.GenerateSubject(models.EmailTypeUpdate, templateData)

		emailReq := &EmailRequest{
			ToEmail:         member.Email,
			ToName:          member.Name,
			Subject:         subject,
			HTMLContent:     htmlContent,
			TextContent:     textContent,
			EmailType:       models.EmailTypeUpdate,
			PregnancyID:     pregnancy.ID,
			VillageMemberID: member.ID,
			UpdateID:        &update.ID,
		}

		err = e.SendEmail(ctx, emailReq)
		if err != nil {
			log.Printf("Failed to send update notification to %s: %v", member.Email, err)
			continue
		}

		log.Printf("Update notification sent to %s for pregnancy %d", member.Email, pregnancy.ID)
	}

	return nil
}

// SendMilestoneNotification sends email notifications for milestone achievements
func (e *EmailService) SendMilestoneNotification(ctx context.Context, milestone *models.PregnancyMilestone, pregnancy *models.Pregnancy) error {
	if !e.config.EmailEnabled {
		log.Printf("Email disabled, skipping milestone notification for pregnancy %d", pregnancy.ID)
		return nil
	}

	// Get village members for this pregnancy
	villageMembers, err := e.getVillageMembers(pregnancy.ID)
	if err != nil {
		return fmt.Errorf("failed to get village members: %w", err)
	}

	if len(villageMembers) == 0 {
		log.Printf("No village members found for pregnancy %d", pregnancy.ID)
		return nil
	}

	// Generate timeline URL
	timelineURL := fmt.Sprintf("%s/view/%s", e.getBaseURL(), pregnancy.ShareID)

	// Prepare template data
	templateData := &TemplateData{
		SenderName:       e.config.SenderName,
		PregnancyID:      pregnancy.ID,
		ParentNames:      e.getParentNames(pregnancy),
		DueDate:          pregnancy.DueDate.Format("January 2, 2006"),
		CurrentWeek:      pregnancy.GetCurrentWeek(),
		TimelineURL:      timelineURL,
		Milestone:        milestone,
		MilestoneTitle:   milestone.Title,
		MilestoneWeek:    milestone.Week,
		MilestoneDate:    milestone.Date.Format("January 2, 2006"),
		MilestoneType:    milestone.GetDisplayType(),
	}

	// Generate email content
	htmlContent, textContent, err := e.MilestoneNotificationTemplate(templateData)
	if err != nil {
		return fmt.Errorf("failed to generate milestone email template: %w", err)
	}

	// Send emails to all village members
	for _, member := range villageMembers {
		templateData.RecipientName = member.Name
		subject := e.GenerateSubject(models.EmailTypeMilestone, templateData)

		emailReq := &EmailRequest{
			ToEmail:         member.Email,
			ToName:          member.Name,
			Subject:         subject,
			HTMLContent:     htmlContent,
			TextContent:     textContent,
			EmailType:       models.EmailTypeMilestone,
			PregnancyID:     pregnancy.ID,
			VillageMemberID: member.ID,
			MilestoneID:     &milestone.ID,
		}

		err = e.SendEmail(ctx, emailReq)
		if err != nil {
			log.Printf("Failed to send milestone notification to %s: %v", member.Email, err)
			continue
		}

		log.Printf("Milestone notification sent to %s for pregnancy %d", member.Email, pregnancy.ID)
	}

	return nil
}

// SendWelcomeEmail sends a welcome email to new village members
func (e *EmailService) SendWelcomeEmail(ctx context.Context, member *models.VillageMember, pregnancy *models.Pregnancy) error {
	if !e.config.EmailEnabled {
		log.Printf("Email disabled, skipping welcome email for member %d", member.ID)
		return nil
	}

	// Generate timeline URL
	baseURL := e.getBaseURL()
	timelineURL := fmt.Sprintf("%s/view/%s", baseURL, pregnancy.ShareID)
	log.Printf("Using BASE_URL: %s for pregnancy %d", baseURL, pregnancy.ID)

	// Generate cover photo URL
	coverPhotoURL := ""
	if pregnancy.CoverPhotoFilename != nil && *pregnancy.CoverPhotoFilename != "" {
		coverPhotoURL = fmt.Sprintf("%s/images/covers/%s", baseURL, *pregnancy.CoverPhotoFilename)
		log.Printf("Generated cover photo URL: %s", coverPhotoURL)
	} else {
		log.Printf("No cover photo found for pregnancy %d - CoverPhotoFilename: %v", pregnancy.ID, pregnancy.CoverPhotoFilename)
	}

	// Prepare template data
	templateData := &TemplateData{
		SenderName:        e.config.SenderName,
		RecipientName:     member.Name,
		PregnancyID:       pregnancy.ID,
		ParentNames:       e.getParentNames(pregnancy),
		DueDate:           pregnancy.DueDate.Format("January 2, 2006"),
		CurrentWeek:       pregnancy.GetCurrentWeek(),
		TimelineURL:       timelineURL,
		CoverPhotoURL:     coverPhotoURL,
		VillageMemberName: member.Name,
	}

	// Generate email content
	htmlContent, textContent, err := e.WelcomeEmailTemplate(templateData)
	if err != nil {
		return fmt.Errorf("failed to generate welcome email template: %w", err)
	}

	subject := e.GenerateSubject(models.EmailTypeWelcome, templateData)

	emailReq := &EmailRequest{
		ToEmail:         member.Email,
		ToName:          member.Name,
		Subject:         subject,
		HTMLContent:     htmlContent,
		TextContent:     textContent,
		EmailType:       models.EmailTypeWelcome,
		PregnancyID:     pregnancy.ID,
		VillageMemberID: member.ID,
	}

	err = e.SendEmail(ctx, emailReq)
	if err != nil {
		return fmt.Errorf("failed to send welcome email to %s: %w", member.Email, err)
	}

	log.Printf("Welcome email sent to %s for pregnancy %d", member.Email, pregnancy.ID)
	return nil
}

// SendTestEmail sends a test email to verify configuration
func (e *EmailService) SendTestEmail(ctx context.Context, toEmail, toName string) error {
	templateData := &TemplateData{
		SenderName:    e.config.SenderName,
		RecipientName: toName,
		ParentNames:   "Test Parents",
		DueDate:       time.Now().AddDate(0, 6, 0).Format("January 2, 2006"),
		CurrentWeek:   20,
		TimelineURL:   e.getBaseURL() + "/test",
	}

	htmlContent := fmt.Sprintf(`
	<h1>Test Email from %s</h1>
	<p>Hello %s!</p>
	<p>This is a test email to verify that your email configuration is working correctly.</p>
	<p>If you received this email, your AWS SES setup is functioning properly.</p>
	<p>Test timeline URL: %s</p>
	<p>Sent at: %s</p>
	`, templateData.SenderName, templateData.RecipientName, templateData.TimelineURL, time.Now().Format("January 2, 2006 3:04 PM"))

	textContent := fmt.Sprintf(`Test Email from %s

Hello %s!

This is a test email to verify that your email configuration is working correctly.
If you received this email, your AWS SES setup is functioning properly.

Test timeline URL: %s
Sent at: %s
`, templateData.SenderName, templateData.RecipientName, templateData.TimelineURL, time.Now().Format("January 2, 2006 3:04 PM"))

	emailReq := &EmailRequest{
		ToEmail:         toEmail,
		ToName:          toName,
		Subject:         fmt.Sprintf("Test Email from %s", e.config.SenderName),
		HTMLContent:     htmlContent,
		TextContent:     textContent,
		EmailType:       "test",
		PregnancyID:     0,
		VillageMemberID: 0,
	}

	return e.SendEmail(ctx, emailReq)
}

// Helper functions

func (e *EmailService) getVillageMembers(pregnancyID int) ([]models.VillageMember, error) {
	query := `
		SELECT id, pregnancy_id, name, email, relationship, created_at 
		FROM village_members 
		WHERE pregnancy_id = ? AND email != '' AND email IS NOT NULL
	`
	
	rows, err := db.GetDB().Query(query, pregnancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.VillageMember
	for rows.Next() {
		var member models.VillageMember
		err := rows.Scan(
			&member.ID,
			&member.PregnancyID,
			&member.Name,
			&member.Email,
			&member.Relationship,
			&member.CreatedAt,
		)
		if err != nil {
			continue
		}
		members = append(members, member)
	}

	return members, nil
}

func (e *EmailService) getUpdatePhotoCount(updateID int) int {
	var count int
	query := `SELECT COUNT(*) FROM update_photos WHERE update_id = ?`
	err := db.GetDB().QueryRow(query, updateID).Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (e *EmailService) getUpdatePhotos(updateID int) []models.UpdatePhoto {
	var photos []models.UpdatePhoto
	query := `SELECT id, update_id, filename, original_filename, file_size, caption, sort_order, created_at 
			  FROM update_photos WHERE update_id = ? ORDER BY sort_order`
	rows, err := db.GetDB().Query(query, updateID)
	if err != nil {
		log.Printf("Error getting update photos: %v", err)
		return photos
	}
	defer rows.Close()

	for rows.Next() {
		var photo models.UpdatePhoto
		err := rows.Scan(&photo.ID, &photo.UpdateID, &photo.Filename, &photo.OriginalFilename, 
						&photo.FileSize, &photo.Caption, &photo.SortOrder, &photo.CreatedAt)
		if err != nil {
			log.Printf("Error scanning photo: %v", err)
			continue
		}
		photos = append(photos, photo)
	}
	return photos
}

func (e *EmailService) getParentNames(pregnancy *models.Pregnancy) string {
	// Get the user's name
	var userName string
	query := `SELECT name FROM users WHERE id = ?`
	err := db.GetDB().QueryRow(query, pregnancy.UserID).Scan(&userName)
	if err != nil {
		userName = "Parent"
	}

	// Add partner name if available
	if pregnancy.PartnerName != nil && *pregnancy.PartnerName != "" {
		return userName + " & " + *pregnancy.PartnerName
	}
	
	return userName
}

func (e *EmailService) getBaseURL() string {
	return e.config.BaseURL
}

// getStringValue safely gets string value from pointer
func getStringValue(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}