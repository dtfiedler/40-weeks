package main

import (
	"fmt"
	"log"
	"net/http"
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

// imageHandler serves images from the data/images directory
func imageHandler(w http.ResponseWriter, r *http.Request) {
	// Get the image path from URL
	imagePath := strings.TrimPrefix(r.URL.Path, "/images/")
	if imagePath == "" {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Construct full file path
	fullPath := filepath.Join("data", "images", imagePath)

	// Security check - ensure path doesn't escape data/images directory
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Join("data", "images")) {
		http.Error(w, "Invalid image path", http.StatusBadRequest)
		return
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

// videoHandler serves videos from the data/videos directory
func videoHandler(w http.ResponseWriter, r *http.Request) {
	// Get the video path from URL
	videoPath := strings.TrimPrefix(r.URL.Path, "/videos/")
	if videoPath == "" {
		http.Error(w, "Invalid video path", http.StatusBadRequest)
		return
	}

	// Construct full file path
	fullPath := filepath.Join("data", "videos", videoPath)

	// Security check - ensure path doesn't escape data/videos directory
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, filepath.Join("data", "videos")) {
		http.Error(w, "Invalid video path", http.StatusBadRequest)
		return
	}

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
	
	// Verify the pregnancy exists (optional - could just serve the page and let JS handle errors)
	_, err := handlers.GetPregnancyByShareID(shareID)
	if err != nil {
		http.Error(w, "Timeline not found", http.StatusNotFound)
		return
	}
	
	// Serve the timeline page
	http.ServeFile(w, r, "public/timeline.html")
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
