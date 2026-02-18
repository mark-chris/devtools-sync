package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mark-chris/devtools-sync/server/internal/auth"
	"github.com/mark-chris/devtools-sync/server/internal/database"
	"github.com/mark-chris/devtools-sync/server/internal/middleware"
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

	// Validate database URL SSL configuration
	dbURL := os.Getenv("DATABASE_URL")
	if err := database.ValidateDatabaseURL(dbURL, isDev); err != nil {
		log.Fatalf("Database URL validation failed: %v", err)
	}

	// Parse max body size configuration
	maxBodySize := parseMaxBodySize(os.Getenv("MAX_BODY_SIZE"))
	log.Printf("Request body size limit: %d bytes (%.2f MB)", maxBodySize, float64(maxBodySize)/(1024*1024))

	corsOrigins := parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Create mux and register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	// Apply CORS and body size limit middleware to all requests
	handler := middleware.CORS(corsOrigins)(middleware.MaxBodySize(maxBodySize)(mux))

	mode := "production"
	if isDev {
		mode = "development"
	}
	log.Printf("Server starting in %s mode on port %s", mode, port)
	if len(corsOrigins) > 0 {
		log.Printf("CORS allowed origins: %v", corsOrigins)
	}
	log.Printf("Health endpoint: http://localhost:%s/health", port)

	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls -- TLS termination handled by reverse proxy in production
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

// parseMaxBodySize parses the MAX_BODY_SIZE environment variable
// Supports formats: "10MB", "1024KB", "1048576" (bytes)
// Default: 10MB
func parseMaxBodySize(value string) int64 {
	if value == "" {
		return 10 * 1024 * 1024 // 10MB default
	}

	// Try parsing as plain number (bytes)
	if bytes, err := strconv.ParseInt(value, 10, 64); err == nil {
		return bytes
	}

	// Parse with unit suffix
	var multiplier int64 = 1
	var numStr string

	if len(value) >= 2 {
		suffix := value[len(value)-2:]
		numPart := value[:len(value)-2]

		switch suffix {
		case "KB":
			multiplier = 1024
			numStr = numPart
		case "MB":
			multiplier = 1024 * 1024
			numStr = numPart
		case "GB":
			multiplier = 1024 * 1024 * 1024
			numStr = numPart
		default:
			// Not a recognized suffix, try parsing as bytes
			numStr = value
		}
	} else {
		numStr = value
	}

	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		log.Printf("Warning: Invalid MAX_BODY_SIZE value '%s', using default 10MB", value)
		return 10 * 1024 * 1024
	}

	return num * multiplier
}

// parseCORSOrigins parses the CORS_ALLOWED_ORIGINS environment variable.
// Returns a slice of origin strings. Empty input returns nil.
func parseCORSOrigins(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}
