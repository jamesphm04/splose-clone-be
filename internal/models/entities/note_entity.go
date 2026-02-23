package entities

import (
	"time"

	"gorm.io/gorm"
)

// Note represents a clinical note written by a User for a Patient.
type Note struct {
	ID        string         `gorm:"type:uuid;primaryKey"           json:"id"`
	PatientID string         `gorm:"type:uuid;not null;index"       json:"patientId"`
	UserID    string         `gorm:"type:uuid;not null;index"       json:"userId"`
	Content   string         `gorm:"type:text"                      json:"content"`
	CreatedAt time.Time      `                                      json:"createdAt"`
	UpdatedAt time.Time      `                                      json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                          json:"-"`

	// Associations
	Patient       Patient        `gorm:"foreignKey:PatientID"    json:"-"`
	User          User           `gorm:"foreignKey:UserID"       json:"-"`
	Conversations []Conversation `gorm:"foreignKey:NoteID"       json:"-"`
	Attachments   []Attachment   `gorm:"foreignKey:NoteID"       json:"-"`
}

func (n *Note) BeforeCreate(_ *gorm.DB) error {
	newUUID(&n.ID)
	return nil
}
