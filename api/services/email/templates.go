package email

import (
	"bytes"
	"fmt"
	"html/template"
	"simple-go/api/models"
)

// TemplateData contains data for email templates
type TemplateData struct {
	// Common data
	SenderName    string
	RecipientName string
	PregnancyID   int
	ParentNames   string
	DueDate       string
	CurrentWeek   int
	TimelineURL   string
	CoverPhotoURL string
	
	// Update-specific data
	Update          *models.PregnancyUpdate
	UpdateTitle     string
	UpdateContent   string
	UpdateWeek      int
	UpdateDate      string
	UpdatePhotos    []string
	
	// Milestone-specific data
	Milestone        *models.PregnancyMilestone
	MilestoneTitle   string
	MilestoneWeek    int
	MilestoneDate    string
	MilestoneType    string
	
	// Village-specific data
	VillageMemberName string
	InviteURL         string
}

// UpdateNotificationTemplate generates email content for pregnancy update notifications
func (e *EmailService) UpdateNotificationTemplate(data *TemplateData) (string, string, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Pregnancy Update</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f8f9fa; }
        .container { max-width: 600px; margin: 0 auto; background-color: #141212ff; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 28px; font-weight: 600; }
        .content { padding: 40px 30px; }
        .update-card { border: 1px solid #e9ecef; border-radius: 8px; padding: 25px; margin: 20px 0; background-color: #f8f9fa; }
        .week-badge { background-color: #667eea; color: white; padding: 8px 16px; border-radius: 20px; font-size: 14px; font-weight: 600; display: inline-block; margin-bottom: 15px; }
        .update-title { font-size: 22px; font-weight: 600; color: #333; margin-bottom: 10px; }
        .update-content { font-size: 16px; line-height: 1.7; color: #555; margin-bottom: 20px; }
        .photos-section { margin-top: 20px; }
        .photo-count { color: #667eea; font-weight: 600; }
        .cta-button { display: inline-block; background-color: #667eea; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: 600; margin: 20px 0; }
        .footer { background-color: #f8f9fa; padding: 30px; text-align: center; color: #666; font-size: 14px; border-top: 1px solid #e9ecef; }
        .footer a { color: #667eea; text-decoration: none; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.SenderName}}</h1>
            <p>New pregnancy update from {{.ParentNames}}</p>
        </div>
        
        <div class="content">
            <h2>Hello {{.RecipientName}}!</h2>
            <p>{{.ParentNames}} just shared a new update from their pregnancy journey.</p>
            
            <div class="update-card">
                {{if .UpdateWeek}}<div class="week-badge">Week {{.UpdateWeek}}</div>{{end}}
                <div class="update-title">{{.UpdateTitle}}</div>
                <div class="update-content">{{.UpdateContent}}</div>
                <div class="update-date">Posted on {{.UpdateDate}}</div>
                {{if .UpdatePhotos}}
                <div class="photos-section">
                    <span class="photo-count">📸 {{len .UpdatePhotos}} photo(s) included</span>
                </div>
                {{end}}
            </div>
            
            <p>Due date: <strong>{{.DueDate}}</strong> • Currently at week <strong>{{.CurrentWeek}}</strong></p>
            
            <a href="{{.TimelineURL}}" class="cta-button">View Full Timeline</a>
        </div>
        
        <div class="footer">
            <p>You're receiving this because you're part of {{.ParentNames}}'s pregnancy village.</p>
            <p><a href="{{.TimelineURL}}">View Timeline</a> | <a href="#">Unsubscribe</a></p>
            <p>© 2024 {{.SenderName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	textTemplate := `New Pregnancy Update from {{.ParentNames}}

Hello {{.RecipientName}}!

{{.ParentNames}} just shared a new update from their pregnancy journey.

{{if .UpdateWeek}}Week {{.UpdateWeek}}: {{end}}{{.UpdateTitle}}

{{.UpdateContent}}

Posted on {{.UpdateDate}}
{{if .UpdatePhotos}}📸 {{len .UpdatePhotos}} photo(s) included{{end}}

Currently at week {{.CurrentWeek}}, due {{.DueDate}}.

View the full timeline: {{.TimelineURL}}

---
You're receiving this because you're part of {{.ParentNames}}'s pregnancy village.
© 2024 {{.SenderName}}. All rights reserved.`

	return e.renderTemplate("update-html", htmlTemplate, data), e.renderTemplate("update-text", textTemplate, data), nil
}

// MilestoneNotificationTemplate generates email content for milestone notifications
func (e *EmailService) MilestoneNotificationTemplate(data *TemplateData) (string, string, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pregnancy Milestone Reached</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f8f9fa; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; }
        .header { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 28px; font-weight: 600; }
        .content { padding: 40px 30px; }
        .milestone-card { border: 1px solid #f5576c; border-radius: 8px; padding: 25px; margin: 20px 0; background: linear-gradient(135deg, #ffeef4 0%, #fff5f8 100%); }
        .milestone-badge { background-color: #f5576c; color: white; padding: 8px 16px; border-radius: 20px; font-size: 14px; font-weight: 600; display: inline-block; margin-bottom: 15px; }
        .milestone-title { font-size: 22px; font-weight: 600; color: #333; margin-bottom: 10px; }
        .milestone-week { font-size: 18px; color: #f5576c; font-weight: 600; margin-bottom: 15px; }
        .cta-button { display: inline-block; background-color: #f5576c; color: white; padding: 15px 30px; text-decoration: none; border-radius: 5px; font-weight: 600; margin: 20px 0; }
        .footer { background-color: #f8f9fa; padding: 30px; text-align: center; color: #666; font-size: 14px; border-top: 1px solid #e9ecef; }
        .footer a { color: #f5576c; text-decoration: none; }
        .celebration { font-size: 24px; text-align: center; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 Milestone Reached!</h1>
            <p>{{.ParentNames}} reached a new pregnancy milestone</p>
        </div>
        
        <div class="content">
            <h2>Hello {{.RecipientName}}!</h2>
            <p>Exciting news! {{.ParentNames}} just reached an important milestone in their pregnancy journey.</p>
            
            <div class="milestone-card">
                <div class="milestone-badge">{{.MilestoneType}}</div>
                <div class="milestone-week">Week {{.MilestoneWeek}}</div>
                <div class="milestone-title">{{.MilestoneTitle}}</div>
                <div class="celebration">🎉 ✨ 🎊</div>
            </div>
            
            <p>Due date: <strong>{{.DueDate}}</strong> • Currently at week <strong>{{.CurrentWeek}}</strong></p>
            
            <a href="{{.TimelineURL}}" class="cta-button">View Full Timeline</a>
        </div>
        
        <div class="footer">
            <p>You're receiving this because you're part of {{.ParentNames}}'s pregnancy village.</p>
            <p><a href="{{.TimelineURL}}">View Timeline</a> | <a href="#">Unsubscribe</a></p>
            <p>© 2024 {{.SenderName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	textTemplate := `🎉 Milestone Reached! 🎉

Hello {{.RecipientName}}!

Exciting news! {{.ParentNames}} just reached an important milestone in their pregnancy journey.

Week {{.MilestoneWeek}}: {{.MilestoneTitle}}
Type: {{.MilestoneType}}

Currently at week {{.CurrentWeek}}, due {{.DueDate}}.

View the full timeline: {{.TimelineURL}}

---
You're receiving this because you're part of {{.ParentNames}}'s pregnancy village.
© 2024 {{.SenderName}}. All rights reserved.`

	return e.renderTemplate("milestone-html", htmlTemplate, data), e.renderTemplate("milestone-text", textTemplate, data), nil
}

// WelcomeEmailTemplate generates welcome email content for new village members
func (e *EmailService) WelcomeEmailTemplate(data *TemplateData) (string, string, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to the Pregnancy Village</title>
    <link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600;700;800&family=DM+Serif+Display:ital@0;1&display=swap" rel="stylesheet">
    <style>
        body { 
            font-family: 'Poppins', 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; 
            line-height: 1.6; 
            color: #333; 
            margin: 0; 
            padding: 0; 
            background-color: #f8f9fa; 
        }
        .container { 
            max-width: 600px; 
            margin: 0 auto; 
            background-color: #ffffff; 
            box-shadow: 0 4px 12px 0 rgb(0 0 0 / 0.1);
        }
        .header { 
            background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 50%, #d97706 100%); 
            color: white; 
            padding: 40px 30px; 
            text-align: center; 
            position: relative;
        }
        .header h1 { 
            margin: 0; 
            font-size: 28px; 
            font-weight: 700; 
            font-family: 'DM Serif Display', serif;
            text-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .header p {
            margin: 10px 0 0 0;
            font-size: 16px;
            opacity: 0.95;
            font-weight: 500;
        }
        .cover-photo {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            border: 4px solid rgba(255,255,255,0.8);
            margin: 20px auto 0;
            object-fit: cover;
            box-shadow: 0 4px 12px rgba(0,0,0,0.2);
        }
        .content { 
            padding: 40px 30px; 
        }
        .welcome-card { 
            border: 2px solid #fbbf24; 
            border-radius: 12px; 
            padding: 25px; 
            margin: 25px 0; 
            background: linear-gradient(135deg, #fffbeb 0%, #fef3c7 100%); 
        }
        .welcome-card h3 {
            color: #d97706;
            font-weight: 600;
            font-size: 18px;
            margin-bottom: 15px;
        }
        .welcome-card ul {
            list-style: none;
            padding: 0;
            margin: 0;
        }
        .welcome-card li {
            padding: 8px 0;
            font-weight: 500;
            color: #78350f;
        }
        .cta-button { 
            display: inline-block; 
            background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 50%, #d97706 100%); 
            color: white; 
            padding: 16px 32px; 
            text-decoration: none; 
            border-radius: 8px; 
            font-weight: 600; 
            margin: 25px 0; 
            font-family: 'Poppins', sans-serif;
            font-size: 16px;
            box-shadow: 0 4px 12px rgba(251, 191, 36, 0.3);
            transition: all 0.2s ease;
            text-align: center;
            display: block;
            max-width: 250px;
            margin-left: auto;
            margin-right: auto;
        }
        .cta-button:hover {
            transform: translateY(-1px);
            box-shadow: 0 6px 16px rgba(251, 191, 36, 0.4);
        }
        .pregnancy-details {
            background-color: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
            text-align: center;
            border-left: 4px solid #fbbf24;
        }
        .footer { 
            background-color: #f8f9fa; 
            padding: 30px; 
            text-align: center; 
            color: #666; 
            font-size: 14px; 
            border-top: 1px solid #e9ecef; 
        }
        .footer a { 
            color: #d97706; 
            text-decoration: none; 
            font-weight: 500;
        }
        .footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to {{.ParentNames}}'s Journey! 👶</h1>
            <p>You've been invited to their pregnancy village</p>
            {{if .CoverPhotoURL}}
            <img src="{{.CoverPhotoURL}}" alt="{{.ParentNames}}" class="cover-photo">
            {{end}}
        </div>
        
        <div class="content">
            <h2 style="color: #d97706; font-weight: 600; margin-bottom: 20px;">Hello {{.RecipientName}}!</h2>
            <p style="font-size: 16px; margin-bottom: 25px;">{{.ParentNames}} has invited you to follow their pregnancy journey and be part of their special moments.</p>
            
            <div class="welcome-card">
                <h3>What you can expect:</h3>
                <ul>
                    <li>📸 Weekly pregnancy updates with photos</li>
                    <li>🎯 Important milestone notifications</li>
                    <li>📅 Keep track of the due date progress</li>
                    <li>👥 Connect with other village members</li>
                </ul>
            </div>
            
            <div class="pregnancy-details">
                <p style="margin: 0; font-weight: 600; color: #78350f;">
                    Due date: <strong>{{.DueDate}}</strong> • Currently at week <strong>{{.CurrentWeek}}</strong>
                </p>
            </div>
            
            <a href="{{.TimelineURL}}" class="cta-button">View Their Timeline</a>
        </div>
        
        <div class="footer">
            <p>Thanks for being part of {{.ParentNames}}'s pregnancy village!</p>
            <p><a href="{{.TimelineURL}}">View Timeline</a> | <a href="#">Unsubscribe</a></p>
            <p>© 2024 {{.SenderName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	textTemplate := `Welcome to {{.ParentNames}}'s Journey! 👶

Hello {{.RecipientName}}!

{{.ParentNames}} has invited you to follow their pregnancy journey and be part of their special moments.

What you can expect:
• 📸 Weekly pregnancy updates with photos
• 🎯 Important milestone notifications  
• 📅 Keep track of the due date progress
• 👥 Connect with other village members

Currently at week {{.CurrentWeek}}, due {{.DueDate}}.

View their timeline: {{.TimelineURL}}

---
Thanks for being part of {{.ParentNames}}'s pregnancy village!
© 2024 {{.SenderName}}. All rights reserved.`

	return e.renderTemplate("welcome-html", htmlTemplate, data), e.renderTemplate("welcome-text", textTemplate, data), nil
}

// renderTemplate renders a template with the given data
func (e *EmailService) renderTemplate(name, templateStr string, data *TemplateData) string {
	tmpl, err := template.New(name).Parse(templateStr)
	if err != nil {
		fmt.Printf("Error parsing template %s: %v\n", name, err)
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		fmt.Printf("Error executing template %s: %v\n", name, err)
		return ""
	}

	return buf.String()
}

// GenerateSubject creates appropriate email subjects
func (e *EmailService) GenerateSubject(emailType string, data *TemplateData) string {
	switch emailType {
	case models.EmailTypeUpdate:
		if data.UpdateWeek > 0 {
			return fmt.Sprintf("Week %d Update from %s", data.UpdateWeek, data.ParentNames)
		}
		return fmt.Sprintf("New update from %s", data.ParentNames)
	case models.EmailTypeMilestone:
		return fmt.Sprintf("🎉 Milestone reached: Week %d - %s", data.MilestoneWeek, data.MilestoneTitle)
	case models.EmailTypeWelcome:
		return fmt.Sprintf("Welcome to %s's pregnancy journey!", data.ParentNames)
	case models.EmailTypeAnnouncement:
		return fmt.Sprintf("Important announcement from %s", data.ParentNames)
	case models.EmailTypeReminder:
		return fmt.Sprintf("Weekly reminder from %s", data.ParentNames)
	default:
		return fmt.Sprintf("Update from %s", data.ParentNames)
	}
}
