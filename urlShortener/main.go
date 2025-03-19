package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"urlShortener/data"
	"urlShortener/repository"
	"urlShortener/services"
	"urlShortener/urlshort"
)

func readFile(filepath string) ([]byte, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading the file %w", err)
	}
	return content, nil
}

func main() {
	var filepath string
	var isDB bool
	flag.StringVar(&filepath, "f", "./data.yaml", "Yaml filepath")
	flag.BoolVar(&isDB, "db", false, "Comes from DB")
	flag.Parse()

	mux := defaultMux()

	if isDB {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		db := data.NewSQLiteDB("text.db")
		if err := db.Connect(ctx); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Disconnect(ctx)

		repo := repository.NewSQLiteURLRepository(db.GetConn())
		service := services.NewURLService(&repo)

		// Seed database if needed
		if err := repo.Seed(ctx); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}

		handler := urlshort.NewDBHandler(&service)
		dbHandler, err := handler.Handle(mux)
		if err != nil {
			log.Fatalf("Failed to handle DB request: %v", err)
		}

		log.Println("Starting the server on :8080")
		http.ListenAndServe(":8080", dbHandler)
	}

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the fallback
	file, err := readFile(filepath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	handler := urlshort.NewHandler(filepath, file)

	var finalHandler http.Handler
	if handler == nil {
		finalHandler = mapHandler
	} else {
		fileHandler, err := handler.Handle(mapHandler)
		if err != nil {
			log.Fatalf("Failed to handle file request: %v", err)
		}
		finalHandler = fileHandler
	}
	log.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", finalHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
