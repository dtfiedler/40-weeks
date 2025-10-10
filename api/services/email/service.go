package email

import (
	"context"
	"fmt"
	"log"
	"simple-go/api/config"
	"simple-go/api/db"
	"simple-go/api/models"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// EmailService handles email operations using AWS SES
type EmailService struct {
	sesClient *ses.Client
	config    *config.Config
}

// EmailRequest represents an email to be sent
type EmailRequest struct {
	ToEmail     string
	ToName      string
	Subject     string
	HTMLContent string
	TextContent string
	EmailType   string
	PregnancyID int
	VillageMemberID int
	UpdateID    *int
	MilestoneID *int
}

// NewEmailService creates a new email service instance
func NewEmailService() (*EmailService, error) {
	cfg := config.AppConfig
	
	if !cfg.EmailEnabled {
		log.Println("Email service is disabled")
		return &EmailService{config: cfg}, nil
	}

	// Configure AWS credentials
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	sesClient := ses.NewFromConfig(awsCfg)

	return &EmailService{
		sesClient: sesClient,
		config:    cfg,
	}, nil
}

// SendEmail sends an email using AWS SES
func (e *EmailService) SendEmail(ctx context.Context, req *EmailRequest) error {
	if !e.config.EmailEnabled {
		log.Printf("Email service disabled, would send: %s to %s", req.Subject, req.ToEmail)
		return nil
	}

	if e.sesClient == nil {
		return fmt.Errorf("SES client not initialized")
	}

	// Create the email input
	input := &ses.SendEmailInput{
		Source: aws.String(fmt.Sprintf("%s <%s>", e.config.SenderName, e.config.SenderEmail)),
		Destination: &types.Destination{
			ToAddresses: []string{req.ToEmail},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(req.Subject),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(req.HTMLContent),
				},
				Text: &types.Content{
					Data: aws.String(req.TextContent),
				},
			},
		},
	}

	// Send the email
	result, err := e.sesClient.SendEmail(ctx, input)
	if err != nil {
		// Log the email attempt as failed
		e.logEmailNotification(req, "", models.DeliveryStatusFailed)
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Log the email as sent
	messageID := ""
	if result.MessageId != nil {
		messageID = *result.MessageId
	}
	
	e.logEmailNotification(req, messageID, models.DeliveryStatusSent)
	
	log.Printf("Email sent successfully to %s, MessageID: %s", req.ToEmail, messageID)
	return nil
}

// logEmailNotification records the email attempt in the database
func (e *EmailService) logEmailNotification(req *EmailRequest, messageID, status string) {
	notification := models.EmailNotification{
		PregnancyID:     req.PregnancyID,
		VillageMemberID: req.VillageMemberID,
		UpdateID:        req.UpdateID,
		MilestoneID:     req.MilestoneID,
		EmailType:       req.EmailType,
		Subject:         req.Subject,
		SentAt:          time.Now(),
		DeliveryStatus:  status,
		CreatedAt:       time.Now(),
	}
	
	if messageID != "" {
		notification.SESMessageID = &messageID
	}

	// Insert into database
	query := `
		INSERT INTO email_notifications (
			pregnancy_id, village_member_id, update_id, milestone_id,
			email_type, subject, sent_at, delivery_status, ses_message_id, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := db.GetDB().Exec(query,
		notification.PregnancyID,
		notification.VillageMemberID,
		notification.UpdateID,
		notification.MilestoneID,
		notification.EmailType,
		notification.Subject,
		notification.SentAt,
		notification.DeliveryStatus,
		notification.SESMessageID,
		notification.CreatedAt,
	)
	
	if err != nil {
		log.Printf("Failed to log email notification: %v", err)
	}
}

// TestEmailConfiguration tests the SES configuration
func (e *EmailService) TestEmailConfiguration(ctx context.Context) error {
	if !e.config.EmailEnabled {
		return fmt.Errorf("email service is disabled")
	}

	if e.sesClient == nil {
		return fmt.Errorf("SES client not initialized")
	}

	// Test by getting send quota
	_, err := e.sesClient.GetSendQuota(ctx, &ses.GetSendQuotaInput{})
	if err != nil {
		return fmt.Errorf("failed to test SES configuration: %w", err)
	}

	return nil
}

// GetEmailStatistics returns email sending statistics
func (e *EmailService) GetEmailStatistics(pregnancyID int) (*models.NotificationSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_sent,
			SUM(CASE WHEN delivery_status = 'delivered' THEN 1 ELSE 0 END) as total_delivered,
			SUM(CASE WHEN delivery_status IN ('failed', 'bounced', 'complaint') THEN 1 ELSE 0 END) as total_failed,
			SUM(CASE WHEN delivery_status = 'bounced' THEN 1 ELSE 0 END) as total_bounced
		FROM email_notifications 
		WHERE pregnancy_id = ?
	`
	
	var summary models.NotificationSummary
	err := db.GetDB().QueryRow(query, pregnancyID).Scan(
		&summary.TotalSent,
		&summary.TotalDelivered,
		&summary.TotalFailed,
		&summary.TotalBounced,
	)
	
	if err != nil {
		return nil, fmt.Errorf("failed to get email statistics: %w", err)
	}
	
	// Calculate delivery rate
	if summary.TotalSent > 0 {
		summary.DeliveryRate = float64(summary.TotalDelivered) / float64(summary.TotalSent) * 100
	}
	
	return &summary, nil
}