package main

import (
	"fmt"
	"log"
	"net/http"

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
	http.HandleFunc("/api/pregnancy", middleware.AuthMiddleware(handlers.CreatePregnancyHandler))
	http.HandleFunc("/api/pregnancy/invite-hash", middleware.AuthMiddleware(handlers.GetInviteHashHandler))
	http.HandleFunc("/api/pregnancy/invite/", handlers.GetPregnancyFromInviteHandler)
	http.HandleFunc("/api/pregnancy/join/", handlers.JoinVillageFromInviteHandler)
	http.HandleFunc("/api/village-members", middleware.AuthMiddleware(villageHandler))
	http.HandleFunc("/api/village-members/bulk", middleware.AuthMiddleware(handlers.CreateVillageMembersBulkHandler))
	http.HandleFunc("/api/village-members/", middleware.AuthMiddleware(villageMemberHandler))
	http.HandleFunc("/app", routes.AppPageHandler)
	http.HandleFunc("/dashboard", routes.DashboardHandler)
	http.HandleFunc("/pregnancy-setup", routes.PregnancySetupPageHandler)
	http.HandleFunc("/village-setup", routes.VillageSetupPageHandler)
	http.HandleFunc("/admin", routes.AdminPageHandler)
	http.HandleFunc("/share/", routes.SharePageHandler)

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

// villageMemberHandler routes individual village member requests (for delete)
func villageMemberHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodDelete:
		handlers.DeleteVillageMemberHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}