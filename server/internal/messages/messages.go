package messages

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/gotext/server/internal/models"
)

// MessageService handles message-related operations
type MessageService struct {
	DB       *sql.DB
	Upgrader websocket.Upgrader
	Clients  map[uuid.UUID]map[*websocket.Conn]bool // Map of user ID to their connections
	Spaces   map[uuid.UUID]map[*websocket.Conn]bool // Map of space ID to connections
}

// NewMessageService creates a new message service
func NewMessageService(db *sql.DB) *MessageService {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	return &MessageService{
		DB:       db,
		Upgrader: upgrader,
		Clients:  make(map[uuid.UUID]map[*websocket.Conn]bool),
		Spaces:   make(map[uuid.UUID]map[*websocket.Conn]bool),
	}
}

// SendMessage saves a new message to the database
func (s *MessageService) SendMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req models.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	senderUUID, _ := uuid.Parse(userID.(string))
	messageID := uuid.New()
	now := time.Now()
	isDirectMessage := req.RecipientID != nil && req.SpaceID == nil

	// Validate the message has either a space ID or recipient ID but not both
	if (req.SpaceID == nil && req.RecipientID == nil) || (req.SpaceID != nil && req.RecipientID != nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message must have either a space ID or recipient ID"})
		return
	}

	// If sending to a space, check if the user is a member
	if req.SpaceID != nil {
		var isMember bool
		err := s.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM space_members WHERE space_id = $1 AND user_id = $2)",
			req.SpaceID, senderUUID,
		).Scan(&isMember)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check space membership"})
			return
		}

		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this space"})
			return
		}
	}

	// Save the message to the database
	query := `
		INSERT INTO messages (id, content, sender_id, space_id, recipient_id, is_direct_message, created_at, updated_at, is_edited)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, content, sender_id, space_id, recipient_id, is_direct_message, created_at, updated_at, is_edited
	`

	var message models.Message
	err := s.DB.QueryRow(
		query,
		messageID,
		req.Content,
		senderUUID,
		req.SpaceID,
		req.RecipientID,
		isDirectMessage,
		now,
		now,
		false,
	).Scan(
		&message.ID,
		&message.Content,
		&message.SenderID,
		&message.SpaceID,
		&message.RecipientID,
		&message.IsDirectMessage,
		&message.CreatedAt,
		&message.UpdatedAt,
		&message.IsEdited,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save message"})
		return
	}

	// Get sender's username for the response
	var username string
	err = s.DB.QueryRow("SELECT username FROM users WHERE id = $1", senderUUID).Scan(&username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user details"})
		return
	}

	response := message.ToResponse()
	response.SenderUsername = username

	// Broadcast message to appropriate recipients
	go s.broadcastMessage(response)

	c.JSON(http.StatusCreated, response)
}

// GetMessages retrieves messages for a space or direct conversation
func (s *MessageService) GetMessages(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userUUID, _ := uuid.Parse(userID.(string))
	spaceIDStr := c.Query("space_id")
	recipientIDStr := c.Query("recipient_id")

	// Either space_id or recipient_id must be provided, but not both
	if (spaceIDStr == "" && recipientIDStr == "") || (spaceIDStr != "" && recipientIDStr != "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide either space_id or recipient_id"})
		return
	}

	var query string
	var args []interface{}

	if spaceIDStr != "" {
		// Get messages from a space
		spaceUUID, err := uuid.Parse(spaceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid space ID"})
			return
		}

		// Check if user is a member of the space
		var isMember bool
		err = s.DB.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM space_members WHERE space_id = $1 AND user_id = $2)",
			spaceUUID, userUUID,
		).Scan(&isMember)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check space membership"})
			return
		}

		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not a member of this space"})
			return
		}

		query = `
			SELECT m.id, m.content, m.sender_id, u.username, m.space_id, m.recipient_id, 
			       m.is_direct_message, m.created_at, m.updated_at, m.is_edited
			FROM messages m
			JOIN users u ON m.sender_id = u.id
			WHERE m.space_id = $1
			ORDER BY m.created_at DESC
			LIMIT 50
		`
		args = []interface{}{spaceUUID}
	} else {
		// Get direct messages between two users
		recipientUUID, err := uuid.Parse(recipientIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipient ID"})
			return
		}

		query = `
			SELECT m.id, m.content, m.sender_id, u.username, m.space_id, m.recipient_id, 
			       m.is_direct_message, m.created_at, m.updated_at, m.is_edited
			FROM messages m
			JOIN users u ON m.sender_id = u.id
			WHERE (m.sender_id = $1 AND m.recipient_id = $2) OR (m.sender_id = $2 AND m.recipient_id = $1)
			ORDER BY m.created_at DESC
			LIMIT 50
		`
		args = []interface{}{userUUID, recipientUUID}
	}

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve messages"})
		return
	}
	defer rows.Close()

	messages := []models.MessageResponse{}
	for rows.Next() {
		var msg models.MessageResponse
		var senderUsername string

		err := rows.Scan(
			&msg.ID,
			&msg.Content,
			&msg.SenderID,
			&senderUsername,
			&msg.SpaceID,
			&msg.RecipientID,
			&msg.IsDirectMessage,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&msg.IsEdited,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process messages"})
			return
		}

		msg.SenderUsername = senderUsername
		messages = append(messages, msg)
	}

	// Reverse the order so the oldest messages come first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	c.JSON(http.StatusOK, messages)
}

// WebSocketHandler handles real-time connections for chat
func (s *MessageService) WebSocketHandler(c *gin.Context) {
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

	// Log connection
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	log.Printf("WebSocket connection from user %s", userUUID.String())

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := s.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not upgrade to WebSocket"})
		return
	}

	// Register the user's connection
	if _, exists := s.Clients[userUUID]; !exists {
		s.Clients[userUUID] = make(map[*websocket.Conn]bool)
	}
	s.Clients[userUUID][conn] = true
	log.Printf("User %s registered with WebSocket, total connections: %d", userUUID.String(), len(s.Clients[userUUID]))

	// Send welcome message
	welcomeMsg := struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}{
		Type:    "system",
		Message: "Connected to chat server",
	}
	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}

	// Clean up when the connection closes
	defer func() {
		log.Printf("WebSocket connection closing for user %s", userUUID.String())
		// Remove connection from user's connections
		delete(s.Clients[userUUID], conn)
		if len(s.Clients[userUUID]) == 0 {
			delete(s.Clients, userUUID)
		}

		// Remove connection from all spaces
		for spaceID, conns := range s.Spaces {
			if _, ok := conns[conn]; ok {
				delete(s.Spaces[spaceID], conn)
				if len(s.Spaces[spaceID]) == 0 {
					delete(s.Spaces, spaceID)
				}
				log.Printf("Removed user %s from space %s", userUUID.String(), spaceID.String())
			}
		}

		conn.Close()
	}()

	// Handle WebSocket messages
	for {
		var msg struct {
			Type      string     `json:"type"`
			SpaceID   *uuid.UUID `json:"space_id,omitempty"`
			Subscribe bool       `json:"subscribe,omitempty"`
		}

		err := conn.ReadJSON(&msg)
		if err != nil {
			// Client disconnected or sent invalid data
			log.Printf("WebSocket read error for user %s: %v", userUUID.String(), err)
			break
		}

		log.Printf("Received WebSocket message from user %s: %+v", userUUID.String(), msg)

		// Handle subscription/unsubscription to spaces
		if msg.Type == "subscribe" && msg.SpaceID != nil {
			// First check if the user is a member of the space
			var isMember bool
			err := s.DB.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM space_members WHERE space_id = $1 AND user_id = $2)",
				msg.SpaceID, userUUID,
			).Scan(&isMember)

			if err != nil {
				log.Printf("Error checking space membership: %v", err)
				continue
			}

			if !isMember {
				log.Printf("User %s tried to subscribe to space %s but is not a member", userUUID.String(), msg.SpaceID.String())
				// Silently fail, don't allow non-members to subscribe
				continue
			}

			// Add or remove the connection from the space's subscriptions
			if msg.Subscribe {
				if _, exists := s.Spaces[*msg.SpaceID]; !exists {
					s.Spaces[*msg.SpaceID] = make(map[*websocket.Conn]bool)
				}
				s.Spaces[*msg.SpaceID][conn] = true
				log.Printf("User %s subscribed to space %s", userUUID.String(), msg.SpaceID.String())
				
				// Send confirmation
				conn.WriteJSON(struct {
					Type    string     `json:"type"`
					SpaceID uuid.UUID  `json:"space_id"`
					Status  string     `json:"status"`
				}{
					Type:    "subscribe_confirm",
					SpaceID: *msg.SpaceID,
					Status:  "subscribed",
				})
			} else {
				if conns, exists := s.Spaces[*msg.SpaceID]; exists {
					delete(conns, conn)
					if len(conns) == 0 {
						delete(s.Spaces, *msg.SpaceID)
					}
					log.Printf("User %s unsubscribed from space %s", userUUID.String(), msg.SpaceID.String())
				}
			}
		}
	}
}

// broadcastMessage sends a message to appropriate recipients
func (s *MessageService) broadcastMessage(message models.MessageResponse) {
	// If it's a direct message, send to both the sender and recipient
	if message.IsDirectMessage && message.RecipientID != nil {
		log.Printf("Broadcasting direct message %s from %s to %s", 
			message.ID.String(), message.SenderID.String(), message.RecipientID.String())
		
		// Send to recipient's connections
		if conns, exists := s.Clients[*message.RecipientID]; exists {
			log.Printf("Recipient %s has %d active connections", message.RecipientID.String(), len(conns))
			for conn := range conns {
				if err := conn.WriteJSON(message); err != nil {
					log.Printf("Error sending to recipient: %v", err)
				} else {
					log.Printf("Message sent to recipient %s", message.RecipientID.String())
				}
			}
		} else {
			log.Printf("Recipient %s is not connected", message.RecipientID.String())
		}

		// Send to sender's other connections
		if conns, exists := s.Clients[message.SenderID]; exists {
			log.Printf("Sender %s has %d active connections", message.SenderID.String(), len(conns))
			for conn := range conns {
				if err := conn.WriteJSON(message); err != nil {
					log.Printf("Error sending to sender: %v", err)
				}
			}
		}
	} else if message.SpaceID != nil {
		// Send to all clients subscribed to the space
		log.Printf("Broadcasting space message %s to space %s", 
			message.ID.String(), message.SpaceID.String())
		
		if conns, exists := s.Spaces[*message.SpaceID]; exists {
			subscriberCount := len(conns)
			log.Printf("Space %s has %d subscribers", message.SpaceID.String(), subscriberCount)
			
			successCount := 0
			for conn := range conns {
				if err := conn.WriteJSON(message); err != nil {
					log.Printf("Error broadcasting to space member: %v", err)
				} else {
					successCount++
				}
			}
			log.Printf("Message broadcast complete: %d/%d successful", successCount, subscriberCount)
		} else {
			log.Printf("No active subscribers for space %s", message.SpaceID.String())
		}
	}
}

// RegisterMessageRoutes registers the routes for message management
func RegisterMessageRoutes(router *gin.RouterGroup, service *MessageService) {
	messages := router.Group("/messages")
	{
		messages.POST("/", service.SendMessage)
		messages.GET("/", service.GetMessages)
		router.GET("/ws", service.WebSocketHandler) // WebSocket endpoint
	}
} 