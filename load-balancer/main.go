package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	URL       string
	Pool      chan bool
	isHealthy bool
}

type Servers struct {
	sync.Mutex
	data            map[string]Server
	urls            []string
	nextServerIndex int
}

func newServers() Servers {
	data := map[string]Server{
		"http://localhost:8081": {URL: "http://localhost:8081", Pool: make(chan bool, 5)},
		"http://localhost:8082": {URL: "http://localhost:8082", Pool: make(chan bool, 5)},
	}
	urls := make([]string, len(data))
	for serverURL, _ := range data {
		urls = append(urls, serverURL)
	}

	return Servers{
		data:            data,
		urls:            urls,
		nextServerIndex: 0,
	}
}

func getServerWithCapacity(ctx context.Context, servers *Servers) (string, error) {
	serversLen := len(servers.urls)
	for {
		servers.Lock()
		for i := 0; i < serversLen; i++ {
			index := servers.nextServerIndex
			serverURL := servers.urls[index]
			servers.nextServerIndex = (servers.nextServerIndex + 1) % serversLen

			if len(servers.data[serverURL].Pool) < cap(servers.data[serverURL].Pool) && servers.data[serverURL].isHealthy {
				servers.data[serverURL].Pool <- true
				servers.Unlock()
				return serverURL, nil
			}
		}
		servers.Unlock()

		select {
		case <-time.After(1 * time.Second):
			continue
		case <-ctx.Done():
			return "", fmt.Errorf("timeout waiting for an available server")
		}
	}
}

func releaseCapacity(servers *Servers, serverURL string) {
	// release server capacity
	servers.Lock()
	defer servers.Unlock()
	if _, exist := servers.data[serverURL]; exist {
		<-servers.data[serverURL].Pool
	}
}

func doRequest(ctx context.Context, servers *Servers, w http.ResponseWriter, r *http.Request) (*http.Response, error) {
	start := time.Now()
	log.Printf("Received request: %s %s\n", r.Method, r.URL.Path)
	serverURL, err := getServerWithCapacity(ctx, servers)
	if err != nil {
		return nil, fmt.Errorf("error selecting a server %w", err)
	}
	defer releaseCapacity(servers, serverURL)
	log.Printf("Selected server %s\n", serverURL)

	targetURL := serverURL + r.URL.Path
	newReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)

	if err != nil {
		return nil, fmt.Errorf("error creating request body %w", err)
	}

	//Copy headers into the forwarded request
	for k, values := range r.Header {
		for _, v := range values {
			newReq.Header.Add(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(newReq)

	if err != nil {
		return nil, fmt.Errorf("error sending the request! %w", err)
	}

	log.Printf("Response from %s: status=%d, took=%v\n", serverURL, resp.StatusCode, time.Since(start))
	return resp, nil
}

func verifyServers(ctx context.Context, servers *Servers, interval time.Duration) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			checkServersStatus(ctx, servers)
		case <-ctx.Done():
			return
		}
	}
}

func checkServersStatus(ctx context.Context, servers *Servers) {
	servers.Lock()
	defer servers.Unlock()

	for serverURL, server := range servers.data {
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

		req, err := http.NewRequestWithContext(checkCtx, "HEAD", serverURL, nil)
		if err != nil {
			server := servers.data[serverURL]
			server.isHealthy = false
			servers.data[serverURL] = server
			log.Printf("server %s become unhealthy", serverURL)
			cancel()
			continue
		}

		res, err := http.DefaultClient.Do(req)
		isHealthy := err == nil && res.StatusCode < 500

		if server.isHealthy != isHealthy {
			if isHealthy {
				log.Printf("server %s recovered and is now healthy\n", serverURL)
			} else {
				log.Printf("server %s become unhealthy", serverURL)
			}
		}

		if res != nil {
			res.Body.Close()
		}

		server := servers.data[serverURL]
		server.isHealthy = isHealthy
		servers.data[serverURL] = server

		cancel()
	}
}

func main() {
	log.Println("Starting up load balancer...")
	defer log.Println("Shutting down load balancer")

	servers := newServers()
	log.Printf("Initializing %d backend servers\n", len(servers.data))

	mux := http.NewServeMux()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go verifyServers(ctx, &servers, 6*time.Second)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		resp, err := doRequest(r.Context(), &servers, w, r)
		if err != nil {
			log.Printf("error forwarding the request: %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		//Copy headers into the response
		for k, values := range resp.Header {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	http.ListenAndServe(":80", mux)
}
