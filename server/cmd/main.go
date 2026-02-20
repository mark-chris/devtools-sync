package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

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

	// Parse TLS configuration
	tlsEnabled := os.Getenv("TLS_ENABLED") == "true"
	tlsCfg, certFile, keyFile := loadTLSConfig(
		tlsEnabled,
		os.Getenv("TLS_CERT_FILE"),
		os.Getenv("TLS_KEY_FILE"),
		os.Getenv("TLS_MIN_VERSION"),
	)

	// Create mux and register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	// Apply CORS and body size limit middleware to all requests
	handler := middleware.CORS(corsOrigins)(middleware.MaxBodySize(maxBodySize)(mux))

	// Create server with timeouts
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if tlsCfg != nil {
		srv.TLSConfig = tlsCfg
	}

	mode := "production"
	if isDev {
		mode = "development"
	}
	scheme := "http"
	if tlsEnabled {
		scheme = "https"
	}
	log.Printf("Server starting in %s mode on port %s", mode, port)
	if tlsEnabled {
		log.Printf("TLS enabled (min version: %s)", os.Getenv("TLS_MIN_VERSION"))
	}
	if len(corsOrigins) > 0 {
		log.Printf("CORS allowed origins: %v", corsOrigins)
	}
	log.Printf("Health endpoint: %s://localhost:%s/health", scheme, port)

	// Start server in a goroutine
	go func() {
		var err error
		if tlsEnabled {
			err = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}

// loadTLSConfig validates TLS configuration and returns the tls.Config,
// cert file path, and key file path. Returns nil config if TLS is disabled.
func loadTLSConfig(enabled bool, certFile, keyFile, minVersion string) (*tls.Config, string, string) {
	if !enabled {
		return nil, "", ""
	}

	if certFile == "" {
		log.Fatal("TLS_CERT_FILE is required when TLS_ENABLED=true")
	}
	if keyFile == "" {
		log.Fatal("TLS_KEY_FILE is required when TLS_ENABLED=true")
	}

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("TLS certificate file not found: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("TLS key file not found: %s", keyFile)
	}

	return &tls.Config{
		MinVersion: parseTLSMinVersion(minVersion),
	}, certFile, keyFile
}

// parseTLSMinVersion parses the TLS_MIN_VERSION environment variable.
// Accepts "1.2" or "1.3". Defaults to TLS 1.2.
func parseTLSMinVersion(value string) uint16 {
	switch value {
	case "", "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		log.Fatalf("Invalid TLS_MIN_VERSION %q: must be \"1.2\" or \"1.3\"", value)
		return 0 // unreachable
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
