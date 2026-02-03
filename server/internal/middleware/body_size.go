package middleware

import (
	"fmt"
	"net/http"
)

// MaxBodySize returns middleware that limits request body size
// maxBytes is the maximum allowed body size in bytes
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

			// Call next handler
			next.ServeHTTP(w, r)

			// Check if body size was exceeded
			// This is handled automatically by MaxBytesReader which will
			// return an error when Read() is called on an oversized body
		})
	}
}

// HandleMaxBytesError checks if the error is from MaxBytesReader and returns appropriate response
func HandleMaxBytesError(w http.ResponseWriter, err error, maxBytes int64) bool {
	if err == nil {
		return false
	}

	// Check if error is from MaxBytesReader
	if err.Error() == "http: request body too large" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		fmt.Fprintf(w, `{"error":"Request body too large","max_size_bytes":%d}`, maxBytes)
		return true
	}

	return false
}
