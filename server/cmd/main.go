package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"devtools-sync-server"}`)
}

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)

	log.Printf("Server starting on port %s", port)
	log.Printf("Health endpoint: http://localhost:%s/health", port)

	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls -- TLS termination handled by reverse proxy in production
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
