package middleware

import (
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

// RateLimit returns middleware that rate limits requests by client IP.
// maxAttempts is the maximum number of requests allowed within window.
// Returns 429 Too Many Requests with Retry-After header when exceeded.
func RateLimit(rl *auth.RateLimiter, maxAttempts int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := GetClientIP(r)

			if err := rl.CheckLimit(ip, maxAttempts, window); err != nil {
				log.Printf("Rate limit exceeded for IP %s on %s %s", ip, r.Method, r.URL.Path)
				w.Header().Set("Retry-After", strconv.Itoa(int(window.Seconds())))
				writeJSON(w, http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests, please try again later",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP extracts the client IP address from the request.
// Checks X-Forwarded-For (first IP), X-Real-IP, then RemoteAddr.
func GetClientIP(r *http.Request) string {
	// X-Forwarded-For: client, proxy1, proxy2
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ip := strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
		if ip != "" {
			return ip
		}
	}

	// X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// RemoteAddr — strip port
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port
		return r.RemoteAddr
	}
	return host
}
