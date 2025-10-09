package routes

import (
	"html/template"
	"net/http"
	"os"
	"strings"
	
	"simple-go/api/handlers"
)

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

func ManageVillagePageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/manage-village.html")
}

func ManagePregnancyPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/manage-pregnancy.html")
}

func PublicTimelinePageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "public/timeline.html")
}

func SharePageHandler(w http.ResponseWriter, r *http.Request) {
	// Extract share_id from URL path
	path := strings.TrimPrefix(r.URL.Path, "/share/")
	shareID := path
	
	// Default to serving the static file if no share_id
	if shareID == "" || shareID == "/" {
		http.ServeFile(w, r, "public/join.html")
		return
	}
	
	// Get pregnancy info for dynamic meta tags
	pregnancy, err := handlers.GetPregnancyByShareID(shareID)
	if err != nil {
		http.ServeFile(w, r, "public/join.html")
		return
	}
	
	// Get user info
	user, err := handlers.GetUserByID(pregnancy.UserID)
	if err != nil {
		http.ServeFile(w, r, "public/join.html")
		return
	}
	
	// Build parent names
	parentNames := user.Name
	if pregnancy.PartnerName != nil && *pregnancy.PartnerName != "" {
		parentNames = user.Name + " & " + *pregnancy.PartnerName
	}
	
	// Read the join.html file
	content, err := os.ReadFile("public/join.html")
	if err != nil {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}
	
	// Replace the meta tags with dynamic content
	htmlStr := string(content)
	
	// Build title and description
	title := "Join " + parentNames + "'s Pregnancy Village - 40Weeks"
	description := parentNames + " are expecting! Join their village to follow their pregnancy journey and get updates."
	
	// Replace existing meta tags
	htmlStr = strings.Replace(htmlStr, "<title>Join Pregnancy Village - 40Weeks</title>", 
		"<title>"+template.HTMLEscapeString(title)+"</title>", 1)
	
	// Add Open Graph and Twitter meta tags after the title
	metaTags := `
	<meta property="og:title" content="` + template.HTMLEscapeString(title) + `">
	<meta property="og:description" content="` + template.HTMLEscapeString(description) + `">
	<meta property="og:type" content="website">
	<meta property="og:url" content="` + template.HTMLEscapeString(r.Host + r.URL.Path) + `">
	<meta property="og:image" content="https://` + template.HTMLEscapeString(r.Host) + `/static/og-image-generator.html">
	
	<meta name="twitter:card" content="summary_large_image">
	<meta name="twitter:title" content="` + template.HTMLEscapeString(title) + `">
	<meta name="twitter:description" content="` + template.HTMLEscapeString(description) + `">
	<meta name="twitter:image" content="https://` + template.HTMLEscapeString(r.Host) + `/static/og-image-generator.html">
	
	<meta name="description" content="` + template.HTMLEscapeString(description) + `">`
	
	// Insert meta tags after viewport meta
	htmlStr = strings.Replace(htmlStr, `<meta name="viewport" content="width=device-width, initial-scale=1">`,
		`<meta name="viewport" content="width=device-width, initial-scale=1">`+metaTags, 1)
	
	// Write the response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlStr))
}