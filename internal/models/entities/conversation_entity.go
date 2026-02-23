package entities

import (
	"time"

	"gorm.io/gorm"
)

// Conversation represents an AI chat session tied to a Note.

type Conversation struct {
	ID        string    `gorm:"type:uuid;primaryKey"     json:"id"`
	NoteID    string    `gorm:"type:uuid;not null;index" json:"noteId"`
	CreatedAt time.Time `                                json:"createdAt"`
	UpdatedAt time.Time `                                json:"updatedAt"`

	// Associations
	Note     Note      `gorm:"foreignKey:NoteID"      json:"-"`
	Messages []Message `gorm:"foreignKey:ConversationID" json:"-"`
}

func (c *Conversation) BeforeCreate(_ *gorm.DB) error {
	newUUID(&c.ID)
	return nil
}
