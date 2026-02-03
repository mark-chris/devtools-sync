package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"status":"healthy","service":"devtools-sync-server"}`)
}

func main() {
	// Validate JWT secret before starting server
	jwtSecret := os.Getenv("JWT_SECRET")
	isDev := auth.IsDevelopmentMode()

	if err := auth.ValidateSecret(jwtSecret, isDev); err != nil {
		log.Fatalf("JWT secret validation failed: %v", err)
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)

	mode := "production"
	if isDev {
		mode = "development"
	}
	log.Printf("Server starting in %s mode on port %s", mode, port)
	log.Printf("Health endpoint: http://localhost:%s/health", port)

	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls -- TLS termination handled by reverse proxy in production
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
