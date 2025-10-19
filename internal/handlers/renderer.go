package handlers

import (
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

// Renderer handles template rendering
type Renderer struct {
	templates *template.Template
	logger    *log.Logger
}

// NewRenderer creates a new template renderer
func NewRenderer(logger *log.Logger) (*Renderer, error) {
	// Create template with custom functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"toJSON": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
	}

	// Parse all templates with functions
	tmpl, err := template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Renderer{
		templates: tmpl,
		logger:    logger,
	}, nil
}

// Render renders a template with data
func (r *Renderer) Render(w io.Writer, name string, data interface{}) error {
	// For each page render, parse the specific template with layout
	// This avoids conflicts between templates that define the same blocks
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"toJSON": func(v interface{}) template.JS {
			b, _ := json.Marshal(v)
			return template.JS(b)
		},
	}

	var tmpl *template.Template
	var err error

	// Login page doesn't use layout, others do
	if name == "login.html" {
		tmpl, err = template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/"+name)
	} else {
		// Parse layout and the specific page template
		tmpl, err = template.New("").Funcs(funcMap).ParseFS(templatesFS, "templates/layout.html", "templates/"+name)
	}

	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

// RenderPage renders a page template and handles errors
func (r *Renderer) RenderPage(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := r.Render(w, name, data); err != nil {
		r.logger.Printf("Failed to render template %s: %v", name, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
