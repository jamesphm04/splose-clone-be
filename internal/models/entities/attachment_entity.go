package entities

import (
	"time"

	"gorm.io/gorm"
)

// Attachment stores metadata about a file uploaded to S3.
// The actual binary is stored in S3; only the URL and metadata live in DB.
type Attachment struct {
	ID        string         `gorm:"type:uuid;primaryKey"              json:"id"`
	NoteID    string         `gorm:"type:uuid;index"                   json:"noteId"`
	MessageID string         `gorm:"type:uuid;index"                   json:"messageId"`
	URL       string         `gorm:"not null"                          json:"url"`
	Name      string         `gorm:"not null"                          json:"name"`
	Type      string         `gorm:"type:varchar(100)"                 json:"type"` // MIME type
	Size      int64          `                                         json:"size"` // bytes
	S3Key     string         `gorm:"type:varchar(256);not null;index"  json:"_"`
	CreatedAt time.Time      `                                         json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                             json:"-"`
}

func (a *Attachment) BeforeCreate(_ *gorm.DB) error {
	newUUID(&a.ID)
	return nil
}
