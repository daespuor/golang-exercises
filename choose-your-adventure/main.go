package main

import (
	"daespuor91/choose-your-adventure/internal/handlers"
	"daespuor91/choose-your-adventure/internal/parser"
	"log"
	"net/http"
)

func main() {
	server := http.NewServeMux()

	//Server static files
	server.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	story, err := parser.ParseJSON()
	if err != nil {
		log.Fatalf("Error parsing the JSON file %v", err)
		return
	}
	//	addTmp := handlers.WithTemplate(template.Must(template.New("").Parse("Hello!")))
	handler := handlers.NewHandler(story)
	server.Handle("/", handler)

	log.Println("Server started at port 8080...")
	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatal("Error setting up the server in port 8080")
	}

}
