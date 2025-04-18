package models

import (
	"time"

	"github.com/google/uuid"
)

// Space represents a chat space (room) in the system
type Space struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   uuid.UUID `json:"creator_id"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SpaceResponse is the data structure returned to clients
type SpaceResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   uuid.UUID `json:"creator_id"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	MemberCount int       `json:"member_count,omitempty"`
}

// ToResponse converts a Space to a SpaceResponse
func (s *Space) ToResponse() SpaceResponse {
	return SpaceResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		CreatorID:   s.CreatorID,
		IsPublic:    s.IsPublic,
		CreatedAt:   s.CreatedAt,
	}
}

// CreateSpaceRequest is the data structure for space creation
type CreateSpaceRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// SpaceMember represents a user's membership in a space
type SpaceMember struct {
	SpaceID  uuid.UUID `json:"space_id"`
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"` // e.g., "admin", "member"
	JoinedAt time.Time `json:"joined_at"`
}

// UpdateSpaceRequest is the data structure for updating space details
type UpdateSpaceRequest struct {
	Name        string `json:"name" validate:"omitempty,min=3,max=100"`
	Description string `json:"description"`
	IsPublic    *bool  `json:"is_public"`
} 