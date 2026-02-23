package entities

import (
	"time"

	"gorm.io/gorm"
)

// Prompt stores reusable AI prompt templates belonging to a User.
type Prompt struct {
	ID          string         `gorm:"type:uuid;primaryKey"          json:"id"`
	Name        string         `gorm:"type:varchar(255);not null"    json:"name"`
	Description string         `gorm:"type:text"                     json:"description"`
	Content     string         `gorm:"type:text;not null"            json:"content"`
	UserID      string         `gorm:"type:uuid;not null;index"      json:"userId"`
	CreatedAt   time.Time      `                                     json:"createdAt"`
	UpdatedAt   time.Time      `                                     json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                         json:"-"`

	// Associations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (p *Prompt) BeforeCreate(_ *gorm.DB) error {
	newUUID(&p.ID)
	return nil
}
