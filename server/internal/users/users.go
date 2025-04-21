package users

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/gotext/server/internal/models"
)

// UserService handles user-related operations
type UserService struct {
	DB *sql.DB
}

// NewUserService creates a new UserService
func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		DB: db,
	}
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(id uuid.UUID) (models.User, error) {
	var user models.User
	
	query := `SELECT id, username, email, password_hash, is_email_verified, email_verification_token, created_at, updated_at 
			  FROM users WHERE id = $1`
	
	row := s.DB.QueryRow(query, id)
	
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
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(email string) (models.User, error) {
	var user models.User
	
	query := `SELECT id, username, email, password_hash, is_email_verified, email_verification_token, created_at, updated_at 
			  FROM users WHERE email = $1`
	
	row := s.DB.QueryRow(query, email)
	
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
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("user not found")
		}
		return models.User{}, err
	}

	return user, nil
}

// Create creates a new user
func (s *UserService) Create(user models.User) error {
	// Check if user already exists
	var exists bool
	err := s.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", 
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
	
	_, err = s.DB.Exec(query, 
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

// GetAll returns a list of all users (excluding sensitive data)
func (s *UserService) GetAll() ([]models.UserResponse, error) {
	// Query to get all users from the database
	query := `
		SELECT id, username, email, is_email_verified, created_at, updated_at
		FROM users
		ORDER BY username
	`
	
	rows, err := s.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	users := []models.UserResponse{}
	for rows.Next() {
		var user models.User
		
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.IsEmailVerified,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		users = append(users, user.ToResponse())
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return users, nil
}

// UpdateVerificationStatus updates a user's email verification status
func (s *UserService) UpdateVerificationStatus(id uuid.UUID, isVerified bool) error {
	query := `UPDATE users SET is_email_verified = $1, updated_at = $2 WHERE id = $3`
	_, err := s.DB.Exec(query, isVerified, time.Now(), id)
	return err
} 