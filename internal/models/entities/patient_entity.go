package entities

import (
	"time"

	"gorm.io/gorm"
)

type Gender string

const (
	GenderMale    Gender = "male"
	GenderFemale  Gender = "female"
	GenderOther   Gender = "other"
	GenderUnknown Gender = "unknown"
)

// Patient stores personal patient information, linked to a User.
type Patient struct {
	ID          string         `gorm:"type:uuid;primaryKey"           json:"id"`
	Email       string         `gorm:"uniqueIndex"                    json:"email,omitempty"`
	FirstName   string         `gorm:"not null"                       json:"firstName"`
	LastName    string         `gorm:"not null"                       json:"lastName"`
	PhoneNumber string         `gorm:"type:varchar(30)"               json:"phoneNumber,omitempty"`
	DateOfBirth *time.Time     `                                      json:"dateOfBirth,omitempty"`
	Gender      Gender         `gorm:"type:varchar(10)"               json:"gender,omitempty"`
	FullAddress string         `gorm:"type:text"                      json:"fullAddress,omitempty"`
	UserID      string         `gorm:"type:uuid;not null;index"       json:"userId"`
	CreatedAt   time.Time      `                                      json:"createdAt"`
	UpdatedAt   time.Time      `                                      json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index"                          json:"-"`

	// Associations (not loaded by default)
	User  User   `gorm:"foreignKey:UserID"  json:"-"`
	Notes []Note `gorm:"foreignKey:PatientID" json:"-"`
}

func (p *Patient) BeforeCreate(_ *gorm.DB) error {
	newUUID(&p.ID)
	return nil
}
