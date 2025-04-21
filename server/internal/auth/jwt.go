package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gotext/server/internal/models"
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

// GenerateToken creates a new JWT token for a user
func GenerateToken(user models.User) (string, error) {
	// Get secret key from environment
	secretKey := getSecretKey()

	// Create the claims
	expirationTime := time.Now().Add(DefaultTokenExpiration)
	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
		Email: user.Email,
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

// getSecretKey gets the JWT secret key from environment or returns a default (for development only)
func getSecretKey() string {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		// Warning: In production, always use an environment variable for the secret key
		// This default is for development only
		secretKey = "gotext_development_secret_key"
		fmt.Println("Warning: Using default JWT secret key. Set JWT_SECRET_KEY environment variable for production.")
	}
	return secretKey
} 