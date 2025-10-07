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

func PregnancySetupPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/pregnancy-setup.html")
}

func LegalPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/legal.html")
}

func AdminPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/admin.html")
}

func AppPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/app.html")
}

func VillageSetupPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/village-setup.html")
}