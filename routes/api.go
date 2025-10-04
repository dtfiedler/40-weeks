package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"simple-go/config"
	"simple-go/db"
	"simple-go/middleware"

	"github.com/golang-jwt/jwt/v4"
)

type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "Hello from Go service!",
		Status:  "healthy",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := db.GetAllUsers(10)
	if err != nil {
		log.Printf("Database error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users": users,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims := &middleware.Claims{}
	_, _ = jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	// Get user details from database
	user, err := db.GetUserByUsername(claims.Username)
	if err != nil || user == nil {
		response := map[string]interface{}{
			"username": claims.Username,
			"message":  "This is your profile",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]interface{}{
		"username": user.Username,
		"email":    user.Email,
		"userId":   user.ID,
		"message":  "This is your profile",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}