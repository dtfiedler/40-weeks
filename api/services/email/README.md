# Email Notification Service

This email service provides automated email notifications for pregnancy updates and milestones using AWS Simple Email Service (SES).

## Features

- **Pregnancy Update Notifications**: Automatically sent when new updates are posted
- **Milestone Notifications**: Sent when pregnancy milestones are reached
- **Welcome Emails**: Sent to new village members
- **Professional Email Templates**: Beautiful HTML and text email templates
- **Delivery Tracking**: Track email delivery status using SES webhooks
- **Email Analytics**: View email statistics and delivery rates

## Configuration

Set the following environment variables:

```bash
# Email Service Configuration
EMAIL_ENABLED=true                     # Enable/disable email notifications
AWS_REGION=us-east-1                  # AWS region for SES
AWS_ACCESS_KEY_ID=your_access_key     # AWS access key ID
AWS_SECRET_ACCESS_KEY=your_secret_key # AWS secret access key
SENDER_EMAIL=noreply@40weeks.app      # From email address (must be verified in SES)
SENDER_NAME=40Weeks                   # From name
```

## AWS SES Setup

1. **Verify your domain or email address** in AWS SES
2. **Request production access** if sending to unverified emails
3. **Set up SNS notifications** (optional) for delivery tracking
4. **Configure bounce and complaint handling**

## API Endpoints

### Test Email Configuration
```
GET /api/email/config-test
```
Tests the SES configuration and connectivity.

### Send Test Email
```
POST /api/email/test
Content-Type: application/json

{
    "to_email": "test@example.com",
    "to_name": "Test User"
}
```

### Get Email Notifications
```
GET /api/email/notifications?limit=50&offset=0
```
Returns email notification history for the current user's pregnancy.

### Get Email Statistics
```
GET /api/email/statistics
```
Returns email delivery statistics including sent, delivered, bounced, and failed counts.

### Send Update Notification (Manual)
```
POST /api/email/send-update
Content-Type: application/json

{
    "update_id": 123
}
```
Manually triggers email notification for a specific update.

## Email Types

- `update`: Pregnancy update notifications
- `milestone`: Milestone achievement notifications  
- `welcome`: Welcome emails for new village members
- `announcement`: General announcements
- `reminder`: Weekly/periodic reminders

## Email Templates

The service includes three main email templates:

### Update Notification Template
- Clean, modern design with pregnancy theme
- Displays update title, content, week number, and date
- Shows photo count and provides link to full timeline
- Responsive design for mobile devices

### Milestone Notification Template
- Celebration-themed design with milestone information
- Displays milestone type, week, and achievement details
- Engaging visuals to highlight the achievement

### Welcome Email Template
- Welcoming design for new village members
- Explains what they can expect from notifications
- Provides timeline access and sets expectations

## Integration

### Automatic Notifications

Email notifications are automatically triggered when:

1. **New pregnancy update is posted** (if `is_shared` is true)
2. **Village member is added** (welcome email)
3. **Milestone is reached** (configured separately)

### Manual Notifications

You can also trigger notifications manually via API endpoints for testing or re-sending.

## Error Handling

- Failed emails are logged with detailed error information
- Retry logic for temporary failures
- Delivery status tracking for monitoring
- Graceful degradation when email service is disabled

## Development

When `EMAIL_ENABLED=false`, the service logs email attempts without actually sending them, making it safe for development and testing.

## Monitoring

Monitor email performance using:

- Email statistics API endpoint
- AWS SES sending statistics
- Database email notification logs
- Application server logs

## Security

- Email addresses are validated before sending
- Templates are sanitized to prevent XSS
- SES credentials are stored securely as environment variables
- Email content is generated server-side to prevent tampering