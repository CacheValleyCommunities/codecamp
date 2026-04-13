package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/mail"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
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

const (
	newsletterRateLimitWindow      = time.Minute
	newsletterRateLimitMaxRequests = 5
	newsletterNameMaxLength        = 120
)

var newsletterLimiter = struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}{
	requests: make(map[string][]time.Time),
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file loaded; using existing environment variables")
	}

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
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if strings.TrimSpace(r.FormValue("website")) != "" {
		writeJSONError(w, http.StatusBadRequest, "Invalid submission")
		return
	}

	if !allowNewsletterRequest(clientIP(r)) {
		writeJSONError(w, http.StatusTooManyRequests, "Too many signup attempts. Please try again shortly.")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	name := strings.TrimSpace(r.FormValue("name"))

	if email == "" {
		writeJSONError(w, http.StatusBadRequest, "Email is required")
		return
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		writeJSONError(w, http.StatusBadRequest, "Please enter a valid email address")
		return
	}

	if len(name) > newsletterNameMaxLength {
		writeJSONError(w, http.StatusBadRequest, "Name is too long")
		return
	}

	// Get Mailgun credentials from environment variables
	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")
	mailingListAddress := os.Getenv("MAILGUN_LIST_ADDRESS")

	if mailgunDomain == "" || mailgunAPIKey == "" || mailingListAddress == "" {
		log.Println("Mailgun credentials not configured for newsletter signup")
		writeJSONError(w, http.StatusInternalServerError, "Newsletter signup is currently unavailable")
		return
	}

	// Add subscriber to Mailgun mailing list
	err = addToMailgunList(mailgunAPIKey, mailingListAddress, email, name)
	if err != nil {
		log.Printf("Failed to add subscriber: %v", err)
		writeJSONError(w, http.StatusBadGateway, "Failed to subscribe to newsletter")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Successfully subscribed to Cache Tech Community newsletter!",
	})
}

func addToMailgunList(apiKey, listAddress, email, name string) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	err := addSubscriberToMailgunList(client, apiKey, listAddress, email, name)
	if err == nil {
		return nil
	}

	var mailgunErr *mailgunHTTPError
	if !errors.As(err, &mailgunErr) || mailgunErr.StatusCode != http.StatusNotFound {
		return err
	}

	if createErr := createMailgunList(client, apiKey, listAddress); createErr != nil {
		return createErr
	}

	return addSubscriberToMailgunList(client, apiKey, listAddress, email, name)
}

func addSubscriberToMailgunList(client *http.Client, apiKey, listAddress, email, name string) error {
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

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return &mailgunHTTPError{StatusCode: resp.StatusCode, Body: buf.String()}
	}

	return nil
}

func createMailgunList(client *http.Client, apiKey, listAddress string) error {
	apiURL := "https://api.mailgun.net/v3/lists"

	data := url.Values{}
	data.Set("address", listAddress)
	data.Set("name", "CodeCamp Newsletter")
	data.Set("description", "CodeCamp newsletter updates")
	data.Set("access_level", "readonly")

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.SetBasicAuth("api", apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return &mailgunHTTPError{StatusCode: resp.StatusCode, Body: buf.String()}
	}

	return nil
}

type mailgunHTTPError struct {
	StatusCode int
	Body       string
}

func (e *mailgunHTTPError) Error() string {
	return fmt.Sprintf("mailgun API error: %d - %s", e.StatusCode, e.Body)
}

func allowNewsletterRequest(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}

	now := time.Now()
	cutoff := now.Add(-newsletterRateLimitWindow)

	newsletterLimiter.mu.Lock()
	defer newsletterLimiter.mu.Unlock()

	recent := make([]time.Time, 0, newsletterRateLimitMaxRequests)
	for _, ts := range newsletterLimiter.requests[ip] {
		if ts.After(cutoff) {
			recent = append(recent, ts)
		}
	}

	if len(recent) >= newsletterRateLimitMaxRequests {
		newsletterLimiter.requests[ip] = recent
		return false
	}

	newsletterLimiter.requests[ip] = append(recent, now)
	return true
}

func clientIP(r *http.Request) string {
	forwardedFor := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	return r.RemoteAddr
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSONResponse(w, status, map[string]string{
		"status":  "error",
		"message": message,
	})
}

func writeJSONResponse(w http.ResponseWriter, status int, payload map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}
