package routes

import "net/http"

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/login.html")
}

func RegisterPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/register.html")
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/dashboard.html")
}