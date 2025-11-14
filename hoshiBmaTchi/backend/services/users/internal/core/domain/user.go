package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct{

	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name 		string `gorm:"not null"`
	Username 	string `gorm:"unique;not null"`
	Email 		string `gorm:"unique;not null"`
	Password 	string `gorm:"not null"`

	DateOfBirth     time.Time
	Gender          string
	ProfilePictureURL string

	SubscribedToNewsletter bool `gorm:"default:false"`
	TwoFactorEnabled bool `gorm:"default:false"`

	IsActive bool `gorm:"default:true"` 
	IsBanned bool `gorm:"default:false"`
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate a new UUID (version 4)
	user.ID = uuid.New()

	if user.Username == ""{
		user.Username = user.ID.String()
	}
	return
}