package handlers

import (
	"daespuor91/choose-your-adventure/internal/model"
	"html/template"
	"log"
	"net/http"
)

var tmp *template.Template
var notFoundTmp *template.Template

func init() {
	tmp = template.Must(template.ParseFiles("./internal/templates/story.html"))
	notFoundTmp = template.Must(template.ParseFiles("./internal/templates/404.html"))
}

// Create your own handler
type HandlerOptions func(h *handler)

func WithTemplate(t *template.Template) func(h *handler) {
	return func(h *handler) {
		h.t = t
	}
}

type handler struct {
	s      model.Story
	t      *template.Template
	pathFn func(r *http.Request) string
}

func NewHandler(s model.Story, options ...HandlerOptions) http.Handler {
	currHandler := handler{s: s, t: tmp, pathFn: pathFn}
	for _, opt := range options {
		opt(&currHandler)
	}
	return currHandler
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	arc := h.pathFn(r)
	storyArc, ok := h.s[arc]

	if !ok {
		log.Printf("Page not found for %s", arc)
		if err := notFoundTmp.Execute(w, nil); err != nil {
			http.Error(w, "error rendering not found page", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Serving page for %s...", arc)
	if err := h.t.Execute(w, storyArc); err != nil {
		http.Error(w, "error rendering the template", http.StatusInternalServerError)
	}
}

func pathFn(r *http.Request) string {
	path := r.URL.Path
	if path == "/" {
		return "intro"
	}

	return path[1:]
}
