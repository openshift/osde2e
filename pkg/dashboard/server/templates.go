package server

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"time"
)

//go:embed templates/*.html
var templateFS embed.FS

var funcMap = template.FuncMap{
	"now": time.Now,
	// localTime converts a time.Time to local timezone, formatted as "2006-01-02 15:04 MST"
	"localTime": func(t time.Time) string {
		return t.Local().Format("2006-01-02 15:04 MST")
	},
	// localTimeShort formats as "2006-01-02 15:04" without timezone suffix
	"localDate": func(t time.Time) string {
		return t.Local().Format("2006-01-02")
	},
	// subtract returns a - b (used in templates for failed count = total - passed)
	"subtract": func(a, b int) int {
		return a - b
	},
}

// renderTemplate renders an HTML template with data.
// Each call parses base.html + the requested page file as a fresh template set
// so that {{define "content"}} blocks from different pages don't collide.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"templates/base.html",
		"templates/"+name,
	)
	if err != nil {
		log.Printf("Error loading template %s: %v", name, err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		log.Printf("Error rendering template %s: %v", name, err)
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
	}
}

// PageData represents common data passed to all pages
type PageData struct {
	ActivePage string
	Data       interface{}
}
