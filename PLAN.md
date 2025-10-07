# 40Weeks - Project Development Plan

## Overview
**40Weeks** is a pregnancy tracking and sharing platform that helps expecting parents share their journey with their "village" (friends and family).

## Core Features

### MVP Features
1. **"Telling Your Village"** - Track who knows about the pregnancy and who doesn't
2. **Pregnancy Updates** - Share updates, photos, doctor appointment summaries
3. **Milestone Tracking** - Key appointments, due dates, planned inductions
4. **Village Notifications** - Email updates to family/friends via Amazon SES
5. **Read-only Village Experience** - Family/friends receive updates via email

### Future Features
- Village member comments and encouragement
- Story sharing from village members
- Interactive timeline for village members
- Mobile app companion

## Development Phases

### Phase 1: Database & Core Structure ✅ IN PROGRESS
**Timeline:** Week 1

#### Database Schema
- **pregnancies** - Core pregnancy information
- **village_members** - Friends/family who receive updates
- **pregnancy_updates** - Weekly updates, photos, appointments
- **milestones** - Key pregnancy milestones
- **email_notifications** - Track sent notifications

#### Tasks
- [x] Design and implement database schema
- [x] Create migration files
- [x] Extend user authentication for expecting parents
- [x] Create core data models
- [x] Set up file upload for photos
- [x] Create beautiful marketing landing page

### Phase 2: Village Management
**Timeline:** Week 2

#### "Telling Your Village" Feature
- [ ] Village member management interface
- [ ] Track "told" vs "not told" status
- [ ] Bulk email composition for announcements
- [ ] Beautiful village management dashboard

#### Tasks
- [ ] Create village.html page
- [ ] Village member CRUD operations
- [ ] Status tracking functionality
- [ ] Import contacts feature

### Phase 3: Updates & Sharing
**Timeline:** Week 3

#### Update Creation System
- [ ] Week-by-week update forms
- [ ] Photo upload and gallery system
- [ ] Doctor appointment summary templates
- [ ] Milestone celebration pages

#### Tasks
- [ ] Create updates.html page
- [ ] Create timeline.html page
- [ ] Photo upload and storage
- [ ] Update templates (appointment types)
- [ ] Milestone tracking interface

### Phase 4: Email Notifications
**Timeline:** Week 4

#### Amazon SES Integration
- [ ] Email template system
- [ ] Automated village notifications
- [ ] Beautiful HTML email templates
- [ ] Unsubscribe management

#### Tasks
- [ ] Amazon SES setup and configuration
- [ ] Email template engine
- [ ] Notification scheduling system
- [ ] Email preferences management

### Phase 5: Polish & UX
**Timeline:** Week 5

#### Enhanced User Experience
- [ ] Mobile-responsive design
- [ ] Pregnancy week calculator
- [ ] Enhanced photo galleries
- [ ] Timeline improvements

#### Tasks
- [ ] Mobile optimization
- [ ] Performance improvements
- [ ] User testing and feedback
- [ ] Documentation and onboarding

## Technical Architecture

### Technology Stack
- **Backend:** Go HTTP service
- **Frontend:** HTML + Tailwind CSS + shadcn/ui components
- **Database:** SQLite with golang-migrate
- **Email:** Amazon SES
- **File Storage:** Local filesystem (upgrade to S3 later)
- **Authentication:** JWT tokens

### Project Structure
```
40Weeks/
├── api/
│   ├── models/
│   │   ├── pregnancy.go
│   │   ├── village.go
│   │   ├── update.go
│   │   └── milestone.go
│   ├── handlers/
│   │   ├── pregnancy.go
│   │   ├── village.go
│   │   ├── updates.go
│   │   └── email.go
│   ├── public/
│   │   ├── pregnancy.html     # Pregnancy setup
│   │   ├── village.html       # Village management
│   │   ├── updates.html       # Create updates
│   │   ├── timeline.html      # View timeline
│   │   └── milestones.html    # Milestone tracking
│   ├── templates/
│   │   └── emails/            # Email templates
│   ├── uploads/               # Photo storage
│   └── db/
│       └── migrations/        # Database migrations
```

## Key User Flows

### 1. Initial Setup Flow
1. User signs up/logs in
2. Creates pregnancy profile (due date, partner info)
3. Sets up village (adds family/friends)
4. Chooses who to "tell" initially

### 2. Update Sharing Flow
1. User creates weekly update
2. Adds photos and appointment notes
3. System sends email notifications to village
4. Village members view updates via email

### 3. Milestone Tracking Flow
1. System tracks pregnancy weeks automatically
2. User marks key milestones as complete
3. Special milestone emails sent to village
4. Celebration pages for big moments

## Success Metrics

### MVP Success Criteria
- [ ] 10+ expecting parents using the 40Weeks platform
- [ ] Average of 5+ village members per pregnancy
- [ ] 90%+ email delivery rate
- [ ] Weekly update sharing by 80% of users

### Future Growth Metrics
- Monthly active users
- Village engagement rate
- Photo upload frequency
- Email open/click rates

## Security & Privacy Considerations

### Data Protection
- Secure photo storage and access
- Email address protection
- Private pregnancy information
- GDPR compliance for EU users

### Access Control
- Only parents can manage their pregnancy
- Village members receive read-only access
- Secure unsubscribe mechanisms
- Optional privacy settings

## Launch Strategy

### Beta Launch
- Friends and family testing
- Local parenting groups
- Social media soft launch

### Public Launch
- Parenting blog partnerships
- Social media campaign
- Word-of-mouth referrals
- Pregnancy app store listings