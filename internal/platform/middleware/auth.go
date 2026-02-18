package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/kapella-hub/NexusClaw/internal/platform/crypto"
	"github.com/kapella-hub/NexusClaw/internal/platform/respond"
)

const userIDKey contextKey = "user_id"

// Auth returns a middleware that verifies Bearer tokens from the Authorization
// header using the provided secret. On success, the authenticated user ID is
// stored in the request context (retrievable via GetUserID).
func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if header == "" {
				respond.Error(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				respond.Error(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			subject, err := crypto.VerifyToken(parts[1], secret)
			if err != nil {
				respond.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID retrieves the authenticated user ID from the context.
// Returns an empty string if no user ID is present.
func GetUserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey).(string)
	return id
}
