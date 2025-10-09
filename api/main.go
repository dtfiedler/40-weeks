package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"simple-go/api/config"
	"simple-go/api/db"
	"simple-go/api/handlers"
	"simple-go/api/middleware"
	"simple-go/api/routes"
)

func main() {
	// Initialize database
	if err := db.InitDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.CloseDB()

	// Setup routes with middleware
	setupRoutes()

	port := ":" + config.AppConfig.ServerPort
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Println("Public routes: /health, /login, /register, /api/login, /api/register")
	fmt.Println("Protected routes: /api/users, /api/profile, /api/pregnancy, /api/pregnancy/current, /app, /dashboard, /pregnancy-setup, /village-setup, /admin")
	fmt.Println("Static files: /static/*")
	fmt.Println("Demo credentials: admin/password")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

func setupRoutes() {
	// Public routes (no middleware)
	http.HandleFunc("/health", routes.HealthHandler)
	http.HandleFunc("/login", routes.LoginPageHandler)
	http.HandleFunc("/register", routes.RegisterPageHandler)
	http.HandleFunc("/legal", routes.LegalPageHandler)
	http.HandleFunc("/api/login", routes.LoginHandler)
	http.HandleFunc("/api/register", routes.RegisterHandler)

	// Protected routes (with auth middleware)
	http.HandleFunc("/api/users", middleware.AuthMiddleware(routes.UsersHandler))
	http.HandleFunc("/api/profile", middleware.AuthMiddleware(routes.ProfileHandler))
	http.HandleFunc("/api/pregnancy/current", middleware.AuthMiddleware(handlers.GetPregnancyHandler))
	http.HandleFunc("/api/pregnancy", middleware.AuthMiddleware(pregnancyHandler))
	http.HandleFunc("/api/pregnancy/invite-hash", middleware.AuthMiddleware(handlers.GetInviteHashHandler))
	http.HandleFunc("/api/pregnancy/invite/", handlers.GetPregnancyFromInviteHandler)
	http.HandleFunc("/api/pregnancy/join/", handlers.JoinVillageFromInviteHandler)
	http.HandleFunc("/api/village-members", middleware.AuthMiddleware(villageHandler))
	http.HandleFunc("/api/village-members/bulk", middleware.AuthMiddleware(handlers.CreateVillageMembersBulkHandler))
	http.HandleFunc("/api/village-members/", middleware.AuthMiddleware(villageMemberHandler))
	http.HandleFunc("/api/timeline", middleware.AuthMiddleware(handlers.GetCombinedTimelineHandler))
	http.HandleFunc("/timeline/", handlers.PublicTimelineHandler)
	http.HandleFunc("/api/updates", middleware.AuthMiddleware(updateHandler))
	http.HandleFunc("/api/updates/", middleware.AuthMiddleware(updateDetailHandler))
	http.HandleFunc("/images/", imageHandler)
	http.HandleFunc("/videos/", videoHandler)
	http.HandleFunc("/app", routes.AppPageHandler)
	http.HandleFunc("/dashboard", routes.DashboardHandler)
	http.HandleFunc("/pregnancy-setup", routes.PregnancySetupPageHandler)
	http.HandleFunc("/village-setup", routes.VillageSetupPageHandler)
	http.HandleFunc("/manage/village", routes.ManageVillagePageHandler)
	http.HandleFunc("/manage/pregnancy", routes.ManagePregnancyPageHandler)
	http.HandleFunc("/admin", routes.AdminPageHandler)
	http.HandleFunc("/share/", routes.SharePageHandler)
	http.HandleFunc("/view/", timelinePageHandler)

	// Serve static files from public directory
	fileServer := http.FileServer(http.Dir("public"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Serve landing page for root
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "public/index.html")
	})
}

// pregnancyHandler routes pregnancy requests
func pregnancyHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlers.CreatePregnancyHandler(w, r)
	case "PUT":
		handlers.UpdatePregnancyHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// villageHandler routes village member requests
func villageHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handlers.GetVillageMembersHandler(w, r)
	case http.MethodPost:
		handlers.CreateVillageMemberHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// villageMemberHandler routes individual village member requests (for update and delete)
func villageMemberHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		handlers.UpdateVillageMemberHandler(w, r)
	case http.MethodDelete:
		handlers.DeleteVillageMemberHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// imageHandler serves images from the configured images directory
func imageHandler(w http.ResponseWriter, r *http.Request) {
	// Get the image path from URL
	imagePath := strings.TrimPrefix(r.URL.Path, "/images/")
	if imagePath == "" {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Construct full file path
	fullPath := filepath.Join(config.AppConfig.ImagesDirectory, imagePath)

	// Security check - ensure path doesn't escape images directory
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, config.AppConfig.ImagesDirectory) {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Set cache headers for images (cache for 1 year since they don't change)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Expires", "Thu, 31 Dec 2037 23:55:55 GMT")

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// videoHandler serves videos from the configured videos directory
func videoHandler(w http.ResponseWriter, r *http.Request) {
	// Get the video path from URL
	videoPath := strings.TrimPrefix(r.URL.Path, "/videos/")
	if videoPath == "" {
		http.Error(w, "Invalid video path", http.StatusBadRequest)
		return
	}

	// Construct full file path
	fullPath := filepath.Join(config.AppConfig.VideosDirectory, videoPath)

	// Security check - ensure path doesn't escape videos directory
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, config.AppConfig.VideosDirectory) {
		http.Error(w, "Invalid video path", http.StatusBadRequest)
		return
	}

	// Set cache headers for videos (cache for 1 year since they don't change)
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Expires", "Thu, 31 Dec 2037 23:55:55 GMT")

	// Set appropriate content type for videos
	ext := strings.ToLower(filepath.Ext(cleanPath))
	switch ext {
	case ".mp4":
		w.Header().Set("Content-Type", "video/mp4")
	case ".mov":
		w.Header().Set("Content-Type", "video/quicktime")
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// timelinePageHandler serves the public timeline page for a pregnancy
func timelinePageHandler(w http.ResponseWriter, r *http.Request) {
	// Extract share_id from URL path
	path := strings.TrimPrefix(r.URL.Path, "/view/")
	shareID := path
	
	if shareID == "" {
		http.Error(w, "Invalid timeline link", http.StatusBadRequest)
		return
	}
	
	// Fetch pregnancy data for meta tags
	pregnancy, err := handlers.GetPregnancyByShareID(shareID)
	if err != nil {
		http.Error(w, "Timeline not found", http.StatusNotFound)
		return
	}
	
	// Get parent names for meta tags
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
	
	currentWeek := pregnancy.GetCurrentWeek()
	dueDate := pregnancy.DueDate.Format("January 2, 2006")
	
	// Generate dynamic meta tags (escape HTML to prevent XSS)
	title := html.EscapeString(fmt.Sprintf("Follow %s's journey!", parentNames))
	description := html.EscapeString(fmt.Sprintf("Follow %s's pregnancy journey. Currently at week %d, due %s", parentNames, currentWeek, dueDate))
	
	// Read the HTML template
	htmlContent, err := os.ReadFile("public/timeline.html")
	if err != nil {
		http.Error(w, "Unable to load page", http.StatusInternalServerError)
		return
	}
	
	// Replace placeholders with actual data
	htmlStr := string(htmlContent)
	htmlStr = strings.ReplaceAll(htmlStr, `<title>Pregnancy Timeline - 40Weeks</title>`, fmt.Sprintf(`<title>%s</title>`, title))
	htmlStr = strings.ReplaceAll(htmlStr, `<meta name="description" content="Follow the pregnancy journey">`, fmt.Sprintf(`<meta name="description" content="%s">`, description))
	htmlStr = strings.ReplaceAll(htmlStr, `<meta property="og:title" content="Pregnancy Timeline - 40Weeks">`, fmt.Sprintf(`<meta property="og:title" content="%s">`, title))
	htmlStr = strings.ReplaceAll(htmlStr, `<meta property="og:description" content="Follow the pregnancy journey">`, fmt.Sprintf(`<meta property="og:description" content="%s">`, description))
	htmlStr = strings.ReplaceAll(htmlStr, `<meta name="twitter:title" content="Pregnancy Timeline - 40Weeks">`, fmt.Sprintf(`<meta name="twitter:title" content="%s">`, title))
	htmlStr = strings.ReplaceAll(htmlStr, `<meta name="twitter:description" content="Follow the pregnancy journey">`, fmt.Sprintf(`<meta name="twitter:description" content="%s">`, description))
	
	// Set content type and write response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlStr))
}

// updateHandler routes update requests
func updateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		handlers.CreateUpdateHandler(w, r)
	case http.MethodGet:
		handlers.GetUpdatesHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// updateDetailHandler routes individual update requests (for update and delete)
func updateDetailHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		handlers.UpdateUpdateHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
