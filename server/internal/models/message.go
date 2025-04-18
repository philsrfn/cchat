package models

import (
	"time"

	"github.com/google/uuid"
)

// Message represents a chat message in the system
type Message struct {
	ID             uuid.UUID  `json:"id"`
	Content        string     `json:"content"`
	SenderID       uuid.UUID  `json:"sender_id"`
	SpaceID        *uuid.UUID `json:"space_id,omitempty"`
	RecipientID    *uuid.UUID `json:"recipient_id,omitempty"`
	IsDirectMessage bool       `json:"is_direct_message"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	IsEdited       bool       `json:"is_edited"`
}

// MessageResponse is the data structure returned to clients
type MessageResponse struct {
	ID             uuid.UUID  `json:"id"`
	Content        string     `json:"content"`
	SenderID       uuid.UUID  `json:"sender_id"`
	SenderUsername string     `json:"sender_username,omitempty"`
	SpaceID        *uuid.UUID `json:"space_id,omitempty"`
	RecipientID    *uuid.UUID `json:"recipient_id,omitempty"`
	IsDirectMessage bool       `json:"is_direct_message"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	IsEdited       bool       `json:"is_edited"`
}

// ToResponse converts a Message to a MessageResponse
func (m *Message) ToResponse() MessageResponse {
	return MessageResponse{
		ID:             m.ID,
		Content:        m.Content,
		SenderID:       m.SenderID,
		SpaceID:        m.SpaceID,
		RecipientID:    m.RecipientID,
		IsDirectMessage: m.IsDirectMessage,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		IsEdited:       m.IsEdited,
	}
}

// CreateMessageRequest is the data structure for message creation
type CreateMessageRequest struct {
	Content     string     `json:"content" validate:"required"`
	SpaceID     *uuid.UUID `json:"space_id"`
	RecipientID *uuid.UUID `json:"recipient_id"`
}

// UpdateMessageRequest is the data structure for updating a message
type UpdateMessageRequest struct {
	Content string `json:"content" validate:"required"`
} 