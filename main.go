package main

import (
	"fmt"
	"log"
	"net/http"

	"simple-go/config"
	"simple-go/db"
	"simple-go/middleware"
	"simple-go/routes"
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
	fmt.Println("Protected routes: /api/users, /api/profile, /dashboard")
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
	http.HandleFunc("/api/login", routes.LoginHandler)
	http.HandleFunc("/api/register", routes.RegisterHandler)

	// Protected routes (with auth middleware)
	http.HandleFunc("/api/users", middleware.AuthMiddleware(routes.UsersHandler))
	http.HandleFunc("/api/profile", middleware.AuthMiddleware(routes.ProfileHandler))
	http.HandleFunc("/dashboard", routes.DashboardHandler)

	// Serve static files from public directory
	fileServer := http.FileServer(http.Dir("public"))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Redirect root to login
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	})
}