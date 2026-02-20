package middleware

import "net/http"

// allowedOrigins is the set of origins permitted to make cross-origin requests.
var allowedOrigins = map[string]bool{
	"http://localhost:3000": true,
	"http://localhost:8080": true,
}

// CORS is a Chi-compatible middleware that handles cross-origin resource
// sharing. It validates the request Origin against an allowed list and sets
// the appropriate CORS headers. Preflight OPTIONS requests receive a 204
// response with no body.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
