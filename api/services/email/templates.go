package email

import (
	"bytes"
	"fmt"
	"html/template"
	"simple-go/api/models"
	"strings"
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
	
	// Access request-specific data
	RequestorEmail        string
	RequestorRelationship string
	RequestorMessage      string
	DashboardURL          string
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
            <h2>Hi {{.RecipientName}}!</h2>
            <p>{{.ParentNames}} just shared a new update from their pregnancy.</p>
            
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

	textTemplate := `New post from {{.ParentNames}}

Hi {{.RecipientName}}!

{{.ParentNames}} just shared a new update!

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
        .content { padding: 20px 30px; }
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
            <p>Exciting news! {{.ParentNames}} just reached an important milestone in their pregnancy.</p>
            
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

Exciting news! {{.ParentNames}} just reached an important milestone in their pregnancy.

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
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8">
    <title>Welcome to {{.ParentNames}}'s Pregnancy</title>
    <!--[if mso]>
    <noscript>
        <xml>
            <o:OfficeDocumentSettings>
                <o:AllowPNG/>
                <o:PixelsPerInch>96</o:PixelsPerInch>
            </o:OfficeDocumentSettings>
        </xml>
    </noscript>
    <![endif]-->
    <style type="text/css">
        /* Reset styles */
        body, table, td, p, a, li, blockquote {
            -webkit-text-size-adjust: 100%;
            -ms-text-size-adjust: 100%;
        }
        table, td {
            mso-table-lspace: 0pt;
            mso-table-rspace: 0pt;
        }
        img {
            -ms-interpolation-mode: bicubic;
            border: 0;
            height: auto;
            line-height: 100%;
            outline: none;
            text-decoration: none;
        }
        
        /* Base styles */
        body {
            margin: 0 !important;
            padding: 0 !important;
            background-color: #f8f9fa !important;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif !important;
            font-size: 16px;
            line-height: 1.6;
            color: #333333;
        }
        
        .email-container {
            max-width: 600px;
            margin: 0 auto;
            background-color: #ffffff;
        }
        
        /* Header with gradient */
        .header {
            background: linear-gradient(135deg, #fbbf24 0%, #fbbf24 50%, #f59e0b 100%);
            padding: 40px 30px;
            text-align: center;
        }
        
        .header h1 {
            margin: 0;
            font-size: 28px;
            font-weight: 700;
            color: #ffffff;
            text-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        
        .header p {
            margin: 10px 0 0 0;
            font-size: 16px;
            color: #ffffff;
            opacity: 0.95;
            font-weight: 500;
        }
        
        .cover-photo {
            width: 100%;
            max-width: 400px;
            height: 300px;
            border-radius: 12px;
            margin: 0 auto 15px;
            display: block;
        }
        
        /* Content area */
        .content {
            padding: 40px 30px;
        }
        
        .content h2 {
            color: #d97706;
            font-weight: 600;
            margin-bottom: 20px;
            font-size: 24px;
        }
        
        .content p {
            font-size: 16px;
            margin-bottom: 25px;
            color: #333333;
        }
        
        /* Welcome card */
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
            margin: 0 0 15px 0;
        }
        
        .welcome-card ul {
            list-style: disc;
            padding-left: 20px;
            margin: 0;
        }
        
        .welcome-card li {
            padding: 4px 0;
            font-weight: 500;
            color: #78350f;
            font-size: 15px;
        }
        
        /* Pregnancy details */
        .pregnancy-details {
            margin: 20px 0;
            text-align: center;
        }
        
        .pregnancy-details p {
            margin: 0;
            font-weight: 700;
            color: #333333;
            font-size: 16px;
        }
        
        /* CTA Button */
        .cta-container {
            text-align: center;
            margin: 30px 0;
        }
        
        .cta-button {
            display: inline-block;
            background: linear-gradient(135deg, #fbbf24 0%, #fbbf24 50%, #f59e0b 100%);
            color: #ffffff !important;
            padding: 16px 32px;
            text-decoration: none;
            border-radius: 8px;
            font-weight: 600;
            font-size: 16px;
            box-shadow: 0 4px 12px rgba(251, 191, 36, 0.3);
        }
        
        /* Footer */
        .footer {
            background-color: #f8f9fa;
            padding: 30px;
            text-align: center;
            border-top: 1px solid #e9ecef;
        }
        
        .footer p {
            color: #666666;
            font-size: 14px;
            margin: 5px 0;
        }
        
        .footer a {
            color: #d97706;
            text-decoration: none;
            font-weight: 500;
        }
    </style>
</head>
<body>
    <div class="email-container">
        <!-- Header with gradient -->
        <div class="header">
            <h1>Welcome to {{.ParentNames}}'s Pregnancy!</h1>
            <p>You've been invited to follow their pregnancy</p>
        </div>
        
        <!-- Cover Photo -->
        {{if .CoverPhotoURL}}
        <div style="padding: 30px 30px 0 30px;">
            <img src="{{.CoverPhotoURL}}" alt="{{.ParentNames}}" class="cover-photo" style="width: 100%; max-width: 400px; height: 350px; border-radius: 12px; margin: 0 auto 15px; display: block; object-fit: cover;">
        </div>
        {{end}}
        
        <!-- Main content -->
        <div class="content">
            <h2>Hello {{.RecipientName}}!</h2>
            <p>{{.ParentNames}} {{if contains .ParentNames "&"}}have{{else}}has{{end}} invited you to follow their pregnancy and be part of their special moments.</p>
            
            <!-- What to expect card -->
            <div class="welcome-card">
                <h3>What you can expect:</h3>
                <ul>
                    <li>Pregnancy updates with photos</li>
                    <li>Important milestone notifications</li>
                    <li>Keep track of the due date progress</li>
                    <li>Connect with others following their pregnancy</li>
                </ul>
            </div>
            
            <!-- Pregnancy details -->
            <div class="pregnancy-details">
                <p>Due date: <strong>{{.DueDate}}</strong></p>
                <p>Currently at week <strong>{{.CurrentWeek}}</strong></p>
            </div>
            
            <!-- CTA Button -->
            <div class="cta-container">
                <a href="{{.TimelineURL}}" class="cta-button">View Their Timeline</a>
            </div>
        </div>
        
        <!-- Footer -->
        <div class="footer">
            <p>Thanks for being part of {{.ParentNames}}'s pregnancy village!</p>
            <p><a href="{{.TimelineURL}}">View Timeline</a> | <a href="#">Unsubscribe</a></p>
            <p>© 2024 {{.SenderName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	textTemplate := `Welcome to {{.ParentNames}}'s Pregnancy!

Hi {{.RecipientName}}!

{{.ParentNames}} {{if contains .ParentNames "&"}}have{{else}}has{{end}} invited you to follow their pregnancy and be part of their special moments.

What you can expect:
• Weekly pregnancy updates with photos
• Important milestone notifications  
• Keep track of the due date progress
• Connect with other village members

Currently at week {{.CurrentWeek}}, due {{.DueDate}}.

View their timeline: {{.TimelineURL}}

---
Thanks for being part of {{.ParentNames}}'s pregnancy village!
© 2024 {{.SenderName}}. All rights reserved.`

	return e.renderTemplate("welcome-html", htmlTemplate, data), e.renderTemplate("welcome-text", textTemplate, data), nil
}

// renderTemplate renders a template with the given data
func (e *EmailService) renderTemplate(name, templateStr string, data *TemplateData) string {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"contains": strings.Contains,
	}
	
	tmpl, err := template.New(name).Funcs(funcMap).Parse(templateStr)
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

// AccessRequestNotificationTemplate generates email content for access request notifications
func (e *EmailService) AccessRequestNotificationTemplate(data *TemplateData) (string, string, error) {
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>New Access Request</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background-color: #f8f9fa; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; }
        .header { background: linear-gradient(135deg, #fbbf24 0%, #fbbf24 50%, #f59e0b 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 28px; font-weight: 600; text-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .header p { margin: 10px 0 0 0; font-size: 16px; color: #ffffff; opacity: 0.95; font-weight: 500; }
        .content { padding: 40px 30px; }
        .content h2 { color: #d97706; font-weight: 600; margin-bottom: 20px; font-size: 24px; }
        .request-card { padding: 0; margin: 20px 0; }
        .requester-name { font-size: 22px; font-weight: 600; color: #92400e; margin-bottom: 10px; }
        .request-detail { font-size: 16px; line-height: 1.7; color: #333; margin-bottom: 15px; }
        .request-message { margin: 15px 0; font-style: italic; }
        .cta-container { text-align: center; margin: 30px 0; }
        .cta-button { display: inline-block; background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%); color: #ffffff !important; padding: 15px 30px; text-decoration: none; border-radius: 8px; font-weight: 600; box-shadow: 0 4px 12px rgba(251, 191, 36, 0.3); }
        .footer { background-color: #f8f9fa; padding: 30px; text-align: center; color: #666; font-size: 14px; border-top: 1px solid #e9ecef; }
        .footer a { color: #d97706; text-decoration: none; font-weight: 500; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Access Request</h1>
            <p>Someone wants to follow your pregnancy timeline</p>
        </div>
        
        <div class="content">
            <h2>Hi {{.ParentNames}}!</h2>
            <p>You have a new request from someone who would like to follow your pregnancy timeline.</p>
            
            <div class="request-card">
                <div class="requester-name">{{.VillageMemberName}}</div>
                <div class="request-detail"><strong>Email:</strong> {{.RequestorEmail}}</div>
                <div class="request-detail"><strong>Relationship:</strong> {{.RequestorRelationship}}</div>
                {{if .RequestorMessage}}
                <div class="request-message">
                    "{{.RequestorMessage}}"
                </div>
                {{end}}
            </div>
            
            <p>You can approve or deny this request from your dashboard.</p>
            
            <div class="cta-container">
                <a href="{{.DashboardURL}}" class="cta-button">Review Request</a>
            </div>
        </div>
        
        <div class="footer">
            <p>You can manage access requests and your village members from your dashboard.</p>
            <p><a href="{{.DashboardURL}}">Go to Dashboard</a> | <a href="{{.TimelineURL}}">View Timeline</a></p>
            <p>© 2024 {{.SenderName}}. All rights reserved.</p>
        </div>
    </div>
</body>
</html>`

	textTemplate := `New Access Request

Hi {{.ParentNames}}!

You have a new request from someone who would like to follow your pregnancy timeline.

Requester: {{.VillageMemberName}}
Email: {{.RequestorEmail}}
Relationship: {{.RequestorRelationship}}
{{if .RequestorMessage}}
Message: "{{.RequestorMessage}}"
{{end}}

You can approve or deny this request from your dashboard: {{.DashboardURL}}

View your timeline: {{.TimelineURL}}

---
You can manage access requests and your village members from your dashboard.
© 2024 {{.SenderName}}. All rights reserved.`

	return e.renderTemplate("access-request-html", htmlTemplate, data), e.renderTemplate("access-request-text", textTemplate, data), nil
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
		return fmt.Sprintf("Welcome to %s's pregnancy!", data.ParentNames)
	case models.EmailTypeAnnouncement:
		return fmt.Sprintf("Important announcement from %s", data.ParentNames)
	case models.EmailTypeReminder:
		return fmt.Sprintf("Weekly reminder from %s", data.ParentNames)
	case "access_request":
		return fmt.Sprintf("New access request for your pregnancy timeline")
	default:
		return fmt.Sprintf("Update from %s", data.ParentNames)
	}
}
