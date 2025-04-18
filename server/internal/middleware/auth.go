package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/cchat/server/internal/auth"
	"github.com/google/uuid"
)

// contextKey is a custom type to avoid context key collisions
type contextKey string

const (
	// UserIDKey is the key used to store the user ID in the request context
	UserIDKey contextKey = "user_id"
	// UserEmailKey is the key used to store the user email in the request context
	UserEmailKey contextKey = "user_email"
)

// AuthMiddleware validates the authentication token and adds the user to the request context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First try to get token from cookie
		user, err := auth.ExtractUserFromCookie(r)
		if err == nil {
			// Add user to request context
			ctx := context.WithValue(r.Context(), "user", user)
			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// If no valid cookie, try Authorization header
		tokenString, err := auth.ExtractTokenFromRequest(r)
		if err == nil {
			// Validate the token
			claims, err := auth.ValidateToken(tokenString)
			if err == nil {
				// Get user from database
				user, err := auth.GetUserByEmail(claims.Email)
				if err == nil {
					// Add user to request context
					ctx := context.WithValue(r.Context(), "user", user)
					// Call the next handler with the updated context
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// No valid authentication found
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success":false,"error":"Unauthorized"}`))
	})
}

// OptionalAuthMiddleware tries to authenticate the user but doesn't require it
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get user from cookie
		user, err := auth.ExtractUserFromCookie(r)
		if err == nil {
			// Add user to request context
			ctx := context.WithValue(r.Context(), "user", user)
			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// If no valid cookie, try Authorization header
		tokenString, err := auth.ExtractTokenFromRequest(r)
		if err == nil {
			// Validate the token
			claims, err := auth.ValidateToken(tokenString)
			if err == nil {
				// Get user from database
				user, err := auth.GetUserByEmail(claims.Email)
				if err == nil {
					// Add user to request context
					ctx := context.WithValue(r.Context(), "user", user)
					// Call the next handler with the updated context
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// No valid authentication found, but that's ok for this middleware
		// Continue with no user in context
		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	return userID, nil
}

// GetUserEmailFromContext retrieves the user email from the request context
func GetUserEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(UserEmailKey).(string)
	if !ok {
		return "", errors.New("user email not found in context")
	}
	return email, nil
}

// RequireAuth is a convenient wrapper for routes that require authentication
func RequireAuth(handler http.HandlerFunc) http.Handler {
	return AuthMiddleware(http.HandlerFunc(handler))
} 