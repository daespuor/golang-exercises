package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	dir := http.Dir("static")
	var port string

	flag.StringVar(&port, "port", ":8080", "Backend server port")
	flag.Parse()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(dir)))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "static/index.html")
			return
		}
		fmt.Fprintf(w, "Received Request: %s on %s", r.URL.Path, port)
	})

	mux.HandleFunc("/healthy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Responding healthy!")
	})

	fmt.Printf("Listening on port %s\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		log.Fatalf("Error starting the server %s", err.Error())
	}
}
