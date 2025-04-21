package middleware

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gotext/server/internal/auth"
	"github.com/gotext/server/internal/users"
)

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Key type for context values
type contextKey string

// Context keys
const (
	UserIDKey    contextKey = "userID"
	UserEmailKey contextKey = "userEmail"
)

// AuthMiddleware is the Gin middleware for authentication
type AuthMiddleware struct {
	DB *sql.DB
	UserService *users.UserService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{
		DB: db,
		UserService: users.NewUserService(db),
	}
}

// GinAuthMiddleware authenticates the request
func (m *AuthMiddleware) GinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No Authorization header, try to get token from cookie
			token, err := c.Cookie("token")
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
				c.Abort()
				return
			}
			
			// Validate the token
			claims, err := auth.ValidateToken(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				c.Abort()
				return
			}
			
			// Set user ID in context for later use
			c.Set("userID", claims.Subject)
			c.Set("userEmail", claims.Email)
			c.Next()
			return
		}
		
		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}
		
		tokenString := parts[1]
		
		// Validate the token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}
		
		// Set user ID in context for later use
		c.Set("userID", claims.Subject)
		c.Set("userEmail", claims.Email)
		c.Next()
	}
}

// OptionalGinAuthMiddleware sets user info if authenticated, but doesn't require auth
func (m *AuthMiddleware) OptionalGinAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No Authorization header, try to get token from cookie
			token, err := c.Cookie("token")
			if err == nil {
				// Validate the token
				claims, err := auth.ValidateToken(token)
				if err == nil {
					// Set user ID in context for later use
					c.Set("userID", claims.Subject)
					c.Set("userEmail", claims.Email)
				}
			}
			// Continue regardless of authentication status
			c.Next()
			return
		}
		
		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format but continue anyway (optional auth)
			c.Next()
			return
		}
		
		tokenString := parts[1]
		
		// Validate the token
		claims, err := auth.ValidateToken(tokenString)
		if err == nil {
			// Set user ID in context for later use
			c.Set("userID", claims.Subject)
			c.Set("userEmail", claims.Email)
		}
		
		// Continue regardless of authentication status
		c.Next()
	}
}

// HttpAuthMiddleware is the HTTP middleware for authentication
func HttpAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenString, err := auth.ExtractTokenFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Validate the token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		// Call the next handler with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuthMiddleware sets user info if authenticated, but doesn't require auth
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		tokenString, err := auth.ExtractTokenFromRequest(r)
		if err == nil {
			// Validate the token
			claims, err := auth.ValidateToken(tokenString)
			if err == nil {
				// Add claims to context
				ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
				ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
				
				// Call the next handler with the new context
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// No valid authentication found, but that's ok for this middleware
		// Continue with no user in context
		next.ServeHTTP(w, r)
	})
}

// GetUserIDFromContext retrieves the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDStr, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return uuid.Nil, errors.New("user ID not found in context")
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in context")
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
	return HttpAuthMiddleware(http.HandlerFunc(handler))
} 