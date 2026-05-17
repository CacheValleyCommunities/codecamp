package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/altcha-org/altcha-lib-go"
	"github.com/joho/godotenv"
)

type PageData struct {
	Title        string
	Subtitle     string
	Year         int
	Event        string
	CurrentPage  string
	ContactType  string
	HeroKicker   string
	HeroTitle    string
	HeroSubtitle string
}

// Template functions
var funcMap = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, errors.New("dict: odd number of arguments")
		}
		m := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, errors.New("dict: keys must be strings")
			}
			m[key] = values[i+1]
		}
		return m, nil
	},
}

var commonTemplates = []string{
	"templates/base.html",
	"templates/components/page-hero.html",
	"templates/components/minor-alert.html",
	"templates/components/minor-waiver-link.html",
}

func parsePageTemplates(extra ...string) *template.Template {
	files := append(append([]string{}, commonTemplates...), extra...)
	return template.Must(template.New("").Funcs(funcMap).ParseFiles(files...))
}

const (
	newsletterRateLimitWindow      = time.Minute
	newsletterRateLimitMaxRequests = 5
	newsletterNameMaxLength        = 120

	contactRateLimitWindow      = time.Minute
	contactRateLimitMaxRequests = 5
	contactNameMaxLength        = 120
	contactMessageMaxLength     = 5000
	contactCompanyMaxLength     = 200
	contactRoleMaxLength        = 200
)

var newsletterLimiter = struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}{
	requests: make(map[string][]time.Time),
}

var contactLimiter = struct {
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
	http.HandleFunc("/contact", contactHandler)
	http.HandleFunc("/contact/sponsor", contactHandler)
	http.HandleFunc("/contact/volunteer", contactHandler)
	http.HandleFunc("/contact-submit", contactSubmitHandler)
	http.HandleFunc("/altcha/challenge", altchaChallengeHandler)

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

	tmpl := parsePageTemplates(
		"templates/home.html",
		"templates/components/hero.html",
		"templates/components/about.html",
		"templates/components/events.html",
		"templates/components/sponsors.html",
		"templates/components/community.html",
		"templates/components/newsletter.html",
	)
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

	tmpl := parsePageTemplates("templates/about.html")
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

	tmpl := parsePageTemplates("templates/rules.html")
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

	tmpl := parsePageTemplates("templates/past-winners.html")
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

	tmpl := parsePageTemplates("templates/sponsors.html")
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

	tmpl := parsePageTemplates("templates/volunteers.html")
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data := PageData{
		CurrentPage: "contact",
		Year:        2025,
	}

	switch r.URL.Path {
	case "/contact/sponsor":
		data.ContactType = "sponsor"
		data.Title = "Sponsor Inquiry - CodeCamp: Bridgerland"
		data.HeroKicker = "Partners"
		data.HeroTitle = "Become a sponsor"
		data.HeroSubtitle = "Tell us about your organization and how you'd like to support CodeCamp"
	case "/contact/volunteer":
		data.ContactType = "volunteer"
		data.Title = "Volunteer - CodeCamp: Bridgerland"
		data.HeroKicker = "Community"
		data.HeroTitle = "Volunteer with us"
		data.HeroSubtitle = "Share your interests and how you'd like to help at CodeCamp"
	default:
		data.ContactType = "general"
		data.Title = "Contact Us - CodeCamp: Bridgerland"
		data.HeroKicker = "Get in Touch"
		data.HeroTitle = "Contact us"
		data.HeroSubtitle = "Have a question? Send us a message and we'll get back to you"
	}

	tmpl := parsePageTemplates(
		"templates/contact.html",
		"templates/components/contact-form.html",
	)
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func contactSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := parseSubmissionForm(r); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid form submission")
		return
	}

	if strings.TrimSpace(r.FormValue("website")) != "" {
		writeJSONError(w, http.StatusBadRequest, "Invalid submission")
		return
	}

	if err := verifyAltchaSubmission(r); err != nil {
		log.Printf("ALTCHA verification failed: %v", err)
		writeJSONError(w, http.StatusBadRequest, "Please complete the verification check.")
		return
	}

	if !allowContactRequest(clientIP(r)) {
		writeJSONError(w, http.StatusTooManyRequests, "Too many messages. Please try again shortly.")
		return
	}

	contactType := strings.TrimSpace(r.FormValue("type"))
	switch contactType {
	case "general", "sponsor", "volunteer":
	default:
		writeJSONError(w, http.StatusBadRequest, "Invalid form type")
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	message := strings.TrimSpace(r.FormValue("message"))
	company := strings.TrimSpace(r.FormValue("company"))
	role := strings.TrimSpace(r.FormValue("role"))

	if name == "" {
		writeJSONError(w, http.StatusBadRequest, "Name is required")
		return
	}
	if len(name) > contactNameMaxLength {
		writeJSONError(w, http.StatusBadRequest, "Name is too long")
		return
	}

	if email == "" {
		writeJSONError(w, http.StatusBadRequest, "Email is required")
		return
	}

	parsedEmail, err := mail.ParseAddress(email)
	if err != nil || parsedEmail.Address != email {
		writeJSONError(w, http.StatusBadRequest, "Please enter a valid email address")
		return
	}

	if message == "" {
		writeJSONError(w, http.StatusBadRequest, "Message is required")
		return
	}
	if len(message) > contactMessageMaxLength {
		writeJSONError(w, http.StatusBadRequest, "Message is too long")
		return
	}

	if len(company) > contactCompanyMaxLength {
		writeJSONError(w, http.StatusBadRequest, "Company name is too long")
		return
	}

	if len(role) > contactRoleMaxLength {
		writeJSONError(w, http.StatusBadRequest, "Role is too long")
		return
	}

	mailgunDomain := os.Getenv("MAILGUN_DOMAIN")
	mailgunAPIKey := os.Getenv("MAILGUN_API_KEY")
	contactTo := os.Getenv("CONTACT_TO_EMAIL")
	contactFrom := os.Getenv("CONTACT_FROM_EMAIL")

	if mailgunDomain == "" || mailgunAPIKey == "" || contactTo == "" || contactFrom == "" {
		log.Println("Contact form email not configured")
		writeJSONError(w, http.StatusServiceUnavailable, "Contact form is currently unavailable")
		return
	}

	subjectPrefix := "[CodeCamp Contact]"
	typeLabel := "General inquiry"
	switch contactType {
	case "sponsor":
		subjectPrefix = "[CodeCamp Sponsor]"
		typeLabel = "Sponsorship inquiry"
	case "volunteer":
		subjectPrefix = "[CodeCamp Volunteer]"
		typeLabel = "Volunteer inquiry"
	}

	subject := fmt.Sprintf("%s %s from %s", subjectPrefix, typeLabel, name)

	var body strings.Builder
	fmt.Fprintf(&body, "Type: %s\n", typeLabel)
	fmt.Fprintf(&body, "Name: %s\n", name)
	fmt.Fprintf(&body, "Email: %s\n", email)
	if company != "" {
		fmt.Fprintf(&body, "Company: %s\n", company)
	}
	if role != "" {
		fmt.Fprintf(&body, "Role/Interest: %s\n", role)
	}
	fmt.Fprintf(&body, "Submitted: %s\n\n", time.Now().Format(time.RFC1123))
	body.WriteString("Message:\n")
	body.WriteString(message)

	err = sendMailgunMessage(mailgunDomain, mailgunAPIKey, contactFrom, contactTo, subject, body.String(), email)
	if err != nil {
		log.Printf("Failed to send contact email: %v", err)
		writeJSONError(w, http.StatusBadGateway, "Failed to send message. Please try again.")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Thanks for reaching out! We'll get back to you soon.",
	})
}

func sendMailgunMessage(domain, apiKey, from, to, subject, text, replyTo string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	apiURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", domain)

	data := url.Values{}
	data.Set("from", from)
	data.Set("to", to)
	data.Set("subject", subject)
	data.Set("text", text)
	if replyTo != "" {
		data.Set("h:Reply-To", replyTo)
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, strings.NewReader(data.Encode()))
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

func allowContactRequest(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}

	now := time.Now()
	cutoff := now.Add(-contactRateLimitWindow)

	contactLimiter.mu.Lock()
	defer contactLimiter.mu.Unlock()

	recent := make([]time.Time, 0, contactRateLimitMaxRequests)
	for _, ts := range contactLimiter.requests[ip] {
		if ts.After(cutoff) {
			recent = append(recent, ts)
		}
	}

	if len(recent) >= contactRateLimitMaxRequests {
		contactLimiter.requests[ip] = recent
		return false
	}

	contactLimiter.requests[ip] = append(recent, now)
	return true
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

func altchaChallengeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hmacKey := os.Getenv("ALTCHA_HMAC_KEY")
	if hmacKey == "" {
		http.Error(w, "Verification not configured", http.StatusServiceUnavailable)
		return
	}

	challenge, err := altcha.CreateChallenge(altcha.ChallengeOptions{
		HMACKey:   hmacKey,
		MaxNumber: 100000,
	})
	if err != nil {
		log.Printf("Failed to create ALTCHA challenge: %v", err)
		http.Error(w, "Failed to create challenge", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(challenge); err != nil {
		log.Printf("Failed to encode ALTCHA challenge: %v", err)
	}
}

func verifyAltchaSubmission(r *http.Request) error {
	hmacKey := os.Getenv("ALTCHA_HMAC_KEY")
	if hmacKey == "" {
		return errors.New("ALTCHA_HMAC_KEY not configured")
	}

	payload := strings.TrimSpace(r.FormValue("altcha"))
	if payload == "" {
		return errors.New("altcha payload missing")
	}

	verified, err := altcha.VerifySolution(payload, hmacKey, true)
	if err != nil {
		return err
	}
	if !verified {
		return errors.New("invalid altcha solution")
	}

	return nil
}

func parseSubmissionForm(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		return r.ParseMultipartForm(10 << 20)
	}
	return r.ParseForm()
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
