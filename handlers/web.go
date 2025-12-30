package handlers

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed templates/*
var templatesFS embed.FS

var indexTemplate *template.Template

func init() {
	var err error
	indexTemplate, err = template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		panic("failed to parse index template: " + err.Error())
	}
}

// WebHandler serves the main web UI
func WebHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, "failed to render template", http.StatusInternalServerError)
	}
}
