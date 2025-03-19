package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

func readFile(path string) ([]byte, error) {
	filepath := path
	if path == "/" {
		filepath = "/index.html"
	}
	content, err := os.ReadFile("./www" + filepath)
	if err != nil {
		return nil, fmt.Errorf("error reading the file: %w", err)
	}

	return content, nil
}

func handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatalf("Error closing connection: %v", err)
		}
	}()
	// Stract requested path
	reader := bufio.NewReader(conn)
	req, err := http.ReadRequest(reader)
	if err != nil {
		log.Fatalf("Error reading request: %v", err)
	}

	var res string

	path := req.URL.Path
	content, err := readFile(path)

	if err != nil {
		log.Printf("Error reading path: %v", err)
		res = "HTTP/1.1 404 Not Found\r\nContent-Type: text/plain\r\n\r\n404 - Not Found\r\n"

		_, err = conn.Write([]byte(res))
		if err != nil {
			log.Fatalf("Error writing response: %v", err)
		}
		return
	}

	res = fmt.Sprintf("HTTP/1.1 200 OK\r\n\r\n%s\r\n", string(content))

	_, err = conn.Write([]byte(res))
	if err != nil {
		log.Fatalf("Error writing response: %v", err)
	}
}

func main() {

	listener, err := net.Listen("tcp", ":80")

	if err != nil {
		log.Fatalf("Error starting TCP listener: %v", err)
	}
	defer func() {
		if err := listener.Close(); err != nil {
			log.Fatalf("Error closing listener: %v", err)
		}
	}()

	log.Println("Web Server listening...")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatalf("Error accepting connection: %v", err)
		}

		go handleConnection(conn)
	}

}
