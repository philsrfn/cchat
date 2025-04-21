package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/gotext/server/internal/models"
	"github.com/gotext/server/internal/users"
)

// Claims represents the JWT claims
type Claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

// AuthService handles authentication-related operations
type AuthService struct {
	DB *sql.DB
	UserService *users.UserService
}

// NewAuthService creates a new AuthService
func NewAuthService(db *sql.DB) *AuthService {
	return &AuthService{
		DB: db,
		UserService: users.NewUserService(db),
	}
}

// Register creates a new user account
func (s *AuthService) Register(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists by email
	_, err := s.UserService.GetByEmail(req.Email)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	} else if !errors.Is(err, errors.New("user not found")) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user existence"})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
		return
	}

	userID := uuid.New()
	verificationToken := uuid.New().String()
	now := time.Now()

	// Create user object
	user := models.User{
		ID:                   userID,
		Username:             req.Username,
		Email:                req.Email,
		PasswordHash:         string(hashedPassword),
		IsEmailVerified:      false,
		EmailVerificationToken: verificationToken,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// Create the user
	err = s.UserService.Create(user)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this username or email already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		}
		return
	}

	// TODO: Send verification email

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": userID,
	})
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user by email
	user, err := s.UserService.GetByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := generateToken(user.ID.String(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set cookie with the token
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600 * 24, // 1 day
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set to true in production with HTTPS
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user.ToResponse(),
	})
}

// Logout handles user logout
func (s *AuthService) Logout(c *gin.Context) {
	// Clear the token cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set to true in production with HTTPS
	})

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// ValidateSession validates a session token
func (s *AuthService) ValidateSession(c *gin.Context) {
	// Try to get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	var tokenString string
	
	if authHeader != "" {
		// Check if the header has the Bearer prefix
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}
	} else {
		// Try to get token from cookie
		cookie, err := c.Cookie("token")
		if err != nil || cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authentication token found"})
			return
		}
		tokenString = cookie
	}
	
	// Validate the token
	claims, err := ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}
	
	// Get user by ID from token
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID in token"})
		return
	}
	
	user, err := s.UserService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	
	// Return authenticated user info
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": user.ToResponse(),
	})
}

// GetProfile retrieves the current user's profile
func (s *AuthService) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user by ID
	user, err := s.UserService.GetByID(userUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user.ToResponse())
}

// GetUsers returns a list of all users
func (s *AuthService) GetUsers(c *gin.Context) {
	users, err := s.UserService.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}
	
	c.JSON(http.StatusOK, users)
}

// generateToken creates a new JWT token for the given user
func generateToken(userID, email string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token valid for 24 hours

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
		Email: email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("your_secret_key")) // Use environment variable in production

	return tokenString, err
}

// ExtractUserFromCookie extracts user data from a session cookie
func ExtractUserFromCookie(r *http.Request) (*models.User, error) {
	// This is a placeholder - implement actual cookie handling
	return nil, errors.New("not implemented")
}

// ExtractTokenFromRequest extracts the JWT token from the request header
func ExtractTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header")
	}

	// Check if the header has the Bearer prefix
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}

	return "", errors.New("invalid authorization header format")
}

// ValidateToken validates the JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("your_secret_key"), nil // Use environment variable in production
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
} 