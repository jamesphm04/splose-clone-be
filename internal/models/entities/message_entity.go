package entities

import (
	"time"

	"gorm.io/gorm"
)

type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

// Message stores a single message in a conversation.
type Message struct {
	ID             string         `gorm:"type:uuid;primaryKey"              json:"id"`
	ConversationID string         `gorm:"type:uuid;not null;index"          json:"conversationId"`
	Role           MessageRole    `gorm:"type:varchar(20);not null"         json:"role"`
	Content        string         `gorm:"type:text"                         json:"content"`
	CreatedAt      time.Time      `                                         json:"createdAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index"                             json:"-"`

	// Associations
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"-"`
	Attachments  []Attachment `gorm:"foreignKey:MessageID"      json:"-"`
}

func (m *Message) BeforeCreate(_ *gorm.DB) error {
	newUUID(&m.ID)
	return nil
}
