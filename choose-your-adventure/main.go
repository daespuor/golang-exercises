package main

import (
	"daespuor91/choose-your-adventure/internal/parser"
	"html/template"
	"log"
	"net/http"
	"strings"
)

func getStoryArcKey(path string) string {
	if path == "/" {
		return "intro"
	}

	return strings.TrimPrefix(path, "/")
}

func renderNotFound(w http.ResponseWriter) error {
	tmp, err := template.ParseFiles("./internal/templates/404.html")

	if err != nil {
		log.Printf("Error parsing 404 template! %v", err)
		return err
	}

	tmp.Execute(w, nil)
	return nil
}

func main() {
	server := http.NewServeMux()

	//Server static files
	server.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	server.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stories, err := parser.ParseJSON()
		if err != nil {
			http.Error(w, "Error getting the page", http.StatusInternalServerError)
			return
		}
		arc := getStoryArcKey(r.URL.Path)
		story, ok := stories[arc]

		if !ok {
			log.Printf("Page not found for %s", arc)
			if err := renderNotFound(w); err != nil {
				http.Error(w, "error rendering not found page", http.StatusInternalServerError)
			}
			return
		}

		// Add template
		tmp, err := template.ParseFiles("./internal/templates/story.html")
		if err != nil {
			log.Printf("Error parsing the HTML template for path %s, %v", arc, err)
			http.Error(w, "Error parsing the HTML template", http.StatusInternalServerError)
			return
		}

		log.Printf("Serving page for %s...", arc)
		tmp.Execute(w, story)
	}))

	log.Println("Server started at port 8080...")
	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatal("Error setting up the server in port 8080")
	}

}
