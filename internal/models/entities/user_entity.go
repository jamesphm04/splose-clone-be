package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newUUID(id *string) {
	if *id == "" {
		*id = uuid.NewString()
	}
}

// User represents an authenticated system user
type User struct {
	ID           string         `gorm:"type:uuid;primaryKey"              json:"id"`
	Email        string         `gorm:"uniqueIndex;not null"              json:"email"`
	PasswordHash string         `gorm:"not null"                          json:"-"` // never expose password hash in API responses
	Username     string         `gorm:"not null"                          json:"username"`
	Role         string         `gorm:"type:varchar(50);default:'user'"   json:"role"`
	CreatedAt    time.Time      `                                         json:"createdAt"`
	UpdatedAt    time.Time      `                                         json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index"                             json:"-"`

	// Associations (not loaded by default)
	Patients []Patient `gorm:"foreignKey:UserID" json:"-"`
	Notes    []Note    `gorm:"foreignKey:UserID" json:"-"`
	Prompts  []Prompt  `gorm:"foreignKey:UserID" json:"-"`
}

func (u *User) BeforeCreate(_ *gorm.DB) error {
	newUUID(&u.ID)
	return nil
}
