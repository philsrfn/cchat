package auth

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cchat/server/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// DefaultTokenExpiration is the default time until a token expires (24 hours)
	DefaultTokenExpiration = 24 * time.Hour
)

var (
	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when the token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrInvalidSigningMethod is returned when the token is signed with an invalid method
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)

// Claims represents the JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for a user
func GenerateToken(user models.User) (string, error) {
	// Get secret key from environment
	secretKey := getSecretKey()

	// Create the claims
	expirationTime := time.Now().Add(DefaultTokenExpiration)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}

	// Create the token with the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	// Get secret key from environment
	secretKey := getSecretKey()

	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// Extract the claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	
	return nil, ErrInvalidToken
}

// ExtractTokenFromRequest extracts the JWT token from the Authorization header
func ExtractTokenFromRequest(r *http.Request) (string, error) {
	// Get the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Check the format of the header
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// getSecretKey gets the JWT secret key from environment or returns a default (for development only)
func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		// Warning: In production, always use an environment variable for the secret key
		// This default is for development only
		secretKey = "cchat_development_secret_key"
		fmt.Println("Warning: Using default JWT secret key. Set JWT_SECRET_KEY environment variable for production.")
	}
	return secretKey
} 