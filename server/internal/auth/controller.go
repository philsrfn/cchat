package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/cchat/server/internal/db"
	"github.com/cchat/server/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Cookie constants
const (
	AuthCookieName = "auth_token"
	CookieMaxAge   = 86400 * 7 // 7 days in seconds
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LoginResponse is returned after successful login
type LoginResponse struct {
	Token string             `json:"token"`
	User  models.UserResponse `json:"user"`
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Create the user
	user := models.User{
		ID:                  uuid.New(),
		Username:            req.Username,
		Email:               req.Email,
		PasswordHash:        string(hashedPassword),
		IsEmailVerified:     false, // User needs to verify email
		EmailVerificationToken: uuid.New().String(),
		CreatedAt:           db.CurrentTime(),
		UpdatedAt:           db.CurrentTime(),
	}

	// Store the user in the database
	if err := createUser(user); err != nil {
		if err.Error() == "user already exists" {
			respondWithError(w, http.StatusConflict, "User with this email or username already exists")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// TODO: Send verification email

	// Return user data (excluding sensitive info)
	RespondWithJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: "User registered successfully. Please verify your email.",
		Data:    user.ToResponse(),
	})
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Fetch the user by email
	user, err := getUserByEmail(req.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := GenerateToken(user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Set auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   CookieMaxAge,
		HttpOnly: true,
		Secure:   r.TLS != nil, // Set Secure flag if using HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	// Return token and user data
	RespondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data: LoginResponse{
			Token: token,
			User:  user.ToResponse(),
		},
	})
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AuthCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
	})

	RespondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Logged out successfully",
	})
}

// ValidateAuthHandler checks if the user's authentication is valid
func ValidateAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET method
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the user from the JWT token (which was verified by the middleware)
	user, ok := r.Context().Value("user").(models.User)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Return the user data
	RespondWithJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    user.ToResponse(),
	})
}

// getUserByEmail fetches a user from the database by email
func getUserByEmail(email string) (models.User, error) {
	var user models.User

	query := `SELECT id, username, email, password_hash, is_email_verified, email_verification_token, created_at, updated_at 
			  FROM users WHERE email = $1`
	
	row := db.DB.QueryRow(query, email)
	err := row.Scan(
		&user.ID, 
		&user.Username, 
		&user.Email, 
		&user.PasswordHash,
		&user.IsEmailVerified,
		&user.EmailVerificationToken,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

// GetUserByEmail is an exported version of getUserByEmail for use in middleware
func GetUserByEmail(email string) (models.User, error) {
	return getUserByEmail(email)
}

// createUser stores a new user in the database
func createUser(user models.User) error {
	// Check if user already exists
	var exists bool
	err := db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", 
		user.Email, user.Username).Scan(&exists)
	
	if err != nil {
		return err
	}

	if exists {
		return errors.New("user already exists")
	}

	// Insert new user
	query := `INSERT INTO users 
			  (id, username, email, password_hash, is_email_verified, email_verification_token, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err = db.DB.Exec(query, 
		user.ID, 
		user.Username, 
		user.Email, 
		user.PasswordHash,
		user.IsEmailVerified,
		user.EmailVerificationToken,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

// ExtractUserFromCookie gets the user from the auth cookie
func ExtractUserFromCookie(r *http.Request) (models.User, error) {
	// Get the auth cookie
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil {
		return models.User{}, errors.New("no auth cookie found")
	}

	// Validate the token from cookie
	claims, err := ValidateToken(cookie.Value)
	if err != nil {
		return models.User{}, err
	}

	// Get the user from database
	user, err := getUserByEmail(claims.Email)
	if err != nil {
		return models.User{}, errors.New("user not found")
	}

	return user, nil
}

// IsTokenExpired checks if a token is expired
func IsTokenExpired(tokenString string) bool {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return true
	}

	// Check if token is expired
	exp := time.Unix(claims.ExpiresAt.Unix(), 0)
	return time.Now().After(exp)
}

// RespondWithJSON writes a JSON response
func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondWithError writes an error response
func respondWithError(w http.ResponseWriter, status int, message string) {
	RespondWithJSON(w, status, Response{
		Success: false,
		Error:   message,
	})
} 