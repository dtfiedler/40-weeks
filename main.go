package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Response struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := Response{
		Message: "Hello from Go service!",
		Status:  "healthy",
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}