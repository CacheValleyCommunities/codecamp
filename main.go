package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type PageData struct {
	Title       string
	Subtitle    string
	Year        int
	Event       string
	CurrentPage string
}

// Template functions
var funcMap = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
}

func main() {
	// Serve static files from summer-2025 folder
	fs := http.FileServer(http.Dir("./archive/summer-2025/"))
	http.Handle("/summer-2025/", http.StripPrefix("/summer-2025/", fs))

	// Serve images from summer-2025/images
	imageFs := http.FileServer(http.Dir("./archive/summer-2025/images"))
	http.Handle("/summer-2025/images/", imageFs)

	// Serve other static assets
	staticFs := http.FileServer(http.Dir("./archive/summer-2025/"))
	http.Handle("/images/", staticFs)
	http.Handle("/day-before/", staticFs)

	// Serve public assets (CSS, JS, images)
	publicFs := http.FileServer(http.Dir("./public/"))
	http.Handle("/public/", http.StripPrefix("/public/", publicFs))

	// Route handlers
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/about", aboutHandler)
	http.HandleFunc("/rules", rulesHandler)
	http.HandleFunc("/past-winners", pastWinnersHandler)
	http.HandleFunc("/sponsors", sponsorsHandler)
	http.HandleFunc("/volunteers", volunteersHandler)
	http.HandleFunc("/game-download", gameDownloadHandler)
	http.HandleFunc("/game-download.html", gameDownloadHandler)
	http.HandleFunc("/newsletter-signup", newsletterSignupHandler)

	log.Println("Starting server on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "CodeCamp: Bridgerland",
		Subtitle:    "Building the future, one line at a time",
		Year:        2025,
		Event:       "Thank You for an Amazing CodeCamp 2025!",
		CurrentPage: "home",
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"templates/base.html",
		"templates/home.html",
		"templates/components/hero.html",
		"templates/components/about.html",
		"templates/components/events.html",
		"templates/components/sponsors.html",
		"templates/components/community.html",
		"templates/components/newsletter.html",
	))
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func gameDownloadHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		CurrentPage: "game-download",
		Year:        2025,
	}

	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/game-downloads.html"))
	err := tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		CurrentPage: "about",
		Year:        2025,
	}

	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/about.html"))
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func rulesHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		CurrentPage: "rules",
		Year:        2025,
	}

	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/rules.html"))
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func pastWinnersHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		CurrentPage: "past-winners",
		Year:        2025,
	}

	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/past-winners.html"))
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func sponsorsHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		CurrentPage: "sponsors",
		Year:        2025,
	}

	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/sponsors.html"))
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func volunteersHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		CurrentPage: "volunteers",
		Year:        2025,
	}

	tmpl := template.Must(template.ParseFiles("templates/base.html", "templates/volunteers.html"))
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func newsletterSignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")

	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Get Mailgun credentials from environment variables
	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")
	mailingListAddress := os.Getenv("MAILGUN_LIST_ADDRESS")

	if mailgunDomain == "" || mailgunAPIKey == "" || mailingListAddress == "" {
		log.Println("Mailgun credentials not configured. Email:", email, "Name:", name)
		// For development, just log and return success
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Newsletter signup received!"})
		return
	}

	// Add subscriber to Mailgun mailing list
	err := addToMailgunList(mailgunAPIKey, mailingListAddress, email, name)
	if err != nil {
		log.Printf("Failed to add subscriber: %v", err)
		http.Error(w, "Failed to subscribe to newsletter", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Successfully subscribed to Cache Tech Community newsletter!"})
}

func addToMailgunList(apiKey, listAddress, email, name string) error {
	apiURL := fmt.Sprintf("https://api.mailgun.net/v3/lists/%s/members", listAddress)

	data := url.Values{}
	data.Set("address", email)
	data.Set("name", name)
	data.Set("subscribed", "true")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth("api", apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return fmt.Errorf("mailgun API error: %d - %s", resp.StatusCode, buf.String())
	}

	return nil
}
