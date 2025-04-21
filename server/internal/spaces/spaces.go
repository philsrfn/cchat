package spaces

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/gotext/server/internal/models"
)

// SpaceService handles space-related operations
type SpaceService struct {
	DB *sql.DB
}

// NewSpaceService creates a new space service
func NewSpaceService(db *sql.DB) *SpaceService {
	return &SpaceService{DB: db}
}

// CreateSpace creates a new chat space
func (s *SpaceService) CreateSpace(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	spaceID := uuid.New()
	creatorID, _ := uuid.Parse(userID.(string))
	now := time.Now()

	// Create the space
	query := `
		INSERT INTO spaces (id, name, description, creator_id, is_public, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, description, creator_id, is_public, created_at, updated_at
	`

	var space models.Space
	err := s.DB.QueryRow(
		query,
		spaceID,
		req.Name,
		req.Description,
		creatorID,
		req.IsPublic,
		now,
		now,
	).Scan(
		&space.ID,
		&space.Name,
		&space.Description,
		&space.CreatorID,
		&space.IsPublic,
		&space.CreatedAt,
		&space.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create space"})
		return
	}

	// Add creator as a member with admin role
	memberQuery := `
		INSERT INTO space_members (space_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = s.DB.Exec(
		memberQuery,
		space.ID,
		creatorID,
		"admin",
		now,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add creator as member"})
		return
	}

	c.JSON(http.StatusCreated, space.ToResponse())
}

// GetSpaces returns all spaces the user has access to
func (s *SpaceService) GetSpaces(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, _ := uuid.Parse(userID.(string))

	// Query to get spaces where user is a member or the space is public
	query := `
		SELECT s.id, s.name, s.description, s.creator_id, s.is_public, s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM space_members WHERE space_id = s.id) as member_count
		FROM spaces s
		LEFT JOIN space_members sm ON s.id = sm.space_id
		WHERE sm.user_id = $1 OR s.is_public = true
		GROUP BY s.id
		ORDER BY s.created_at DESC
	`

	rows, err := s.DB.Query(query, userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve spaces"})
		return
	}
	defer rows.Close()

	spaces := []models.SpaceResponse{}
	for rows.Next() {
		var space models.Space
		var response models.SpaceResponse
		var memberCount int

		err := rows.Scan(
			&space.ID,
			&space.Name,
			&space.Description,
			&space.CreatorID,
			&space.IsPublic,
			&space.CreatedAt,
			&space.UpdatedAt,
			&memberCount,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process spaces"})
			return
		}

		response = space.ToResponse()
		response.MemberCount = memberCount
		spaces = append(spaces, response)
	}

	c.JSON(http.StatusOK, spaces)
}

// GetSpaceByID returns details of a specific space
func (s *SpaceService) GetSpaceByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	spaceID := c.Param("id")
	spaceUUID, err := uuid.Parse(spaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	userUUID, _ := uuid.Parse(userID.(string))

	// First check if the user has access to this space
	accessQuery := `
		SELECT EXISTS (
			SELECT 1 FROM spaces s
			LEFT JOIN space_members sm ON s.id = sm.space_id
			WHERE s.id = $1 AND (sm.user_id = $2 OR s.is_public = true)
		)
	`
	var hasAccess bool
	err = s.DB.QueryRow(accessQuery, spaceUUID, userUUID).Scan(&hasAccess)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check space access"})
		return
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have access to this space"})
		return
	}

	// Get space details with member count
	query := `
		SELECT s.id, s.name, s.description, s.creator_id, s.is_public, s.created_at, s.updated_at,
			(SELECT COUNT(*) FROM space_members WHERE space_id = s.id) as member_count
		FROM spaces s
		WHERE s.id = $1
	`

	var space models.Space
	var response models.SpaceResponse
	var memberCount int

	err = s.DB.QueryRow(query, spaceUUID).Scan(
		&space.ID,
		&space.Name,
		&space.Description,
		&space.CreatorID,
		&space.IsPublic,
		&space.CreatedAt,
		&space.UpdatedAt,
		&memberCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve space"})
		}
		return
	}

	response = space.ToResponse()
	response.MemberCount = memberCount

	c.JSON(http.StatusOK, response)
}

// JoinSpace allows a user to join a public space
func (s *SpaceService) JoinSpace(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	spaceID := c.Param("id")
	spaceUUID, err := uuid.Parse(spaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	userUUID, _ := uuid.Parse(userID.(string))

	// Check if the space is public
	var isPublic bool
	err = s.DB.QueryRow("SELECT is_public FROM spaces WHERE id = $1", spaceUUID).Scan(&isPublic)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Space not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check space"})
		}
		return
	}

	if !isPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot join a private space without invitation"})
		return
	}

	// Check if user is already a member
	var isMember bool
	err = s.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM space_members WHERE space_id = $1 AND user_id = $2)",
		spaceUUID, userUUID,
	).Scan(&isMember)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check membership"})
		return
	}

	if isMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You are already a member of this space"})
		return
	}

	// Add user as a member
	_, err = s.DB.Exec(
		"INSERT INTO space_members (space_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)",
		spaceUUID, userUUID, "member", time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join space"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the space"})
}

// InviteToSpace lets an admin invite a user to a space
func (s *SpaceService) InviteToSpace(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	spaceID := c.Param("id")
	spaceUUID, err := uuid.Parse(spaceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
		return
	}

	inviterUUID, _ := uuid.Parse(userID.(string))

	// Verify the inviter is an admin
	var isAdmin bool
	err = s.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM space_members WHERE space_id = $1 AND user_id = $2 AND role = 'admin')",
		spaceUUID, inviterUUID,
	).Scan(&isAdmin)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify permissions"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can invite users"})
		return
	}

	// Parse the invitation request
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the user by email
	var inviteeUUID uuid.UUID
	err = s.DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&inviteeUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
		}
		return
	}

	// Check if user is already a member
	var isMember bool
	err = s.DB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM space_members WHERE space_id = $1 AND user_id = $2)",
		spaceUUID, inviteeUUID,
	).Scan(&isMember)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check membership"})
		return
	}

	if isMember {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a member of this space"})
		return
	}

	// Add user as a member
	_, err = s.DB.Exec(
		"INSERT INTO space_members (space_id, user_id, role, joined_at) VALUES ($1, $2, $3, $4)",
		spaceUUID, inviteeUUID, "member", time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to space"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Successfully invited %s to the space", req.Email)})
}

// RegisterSpaceRoutes registers the routes for space management
func RegisterSpaceRoutes(router *gin.RouterGroup, service *SpaceService) {
	spaces := router.Group("/spaces")
	{
		spaces.POST("/", service.CreateSpace)
		spaces.GET("/", service.GetSpaces)
		spaces.GET("/:id", service.GetSpaceByID)
		spaces.POST("/:id/join", service.JoinSpace)
		spaces.POST("/:id/invite", service.InviteToSpace)
	}
} 