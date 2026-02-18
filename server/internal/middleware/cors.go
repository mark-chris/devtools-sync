package middleware

import "net/http"

// CORS returns middleware that handles Cross-Origin Resource Sharing.
// allowedOrigins is a list of origins permitted to make cross-origin requests.
// If empty, no CORS headers are set (secure by default).
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always set Vary: Origin for correct HTTP caching
			w.Header().Set("Vary", "Origin")

			origin := r.Header.Get("Origin")
			if origin == "" || !originSet[origin] {
				next.ServeHTTP(w, r)
				return
			}

			// Origin is allowed — set CORS response headers
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
