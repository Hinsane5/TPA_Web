package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Follow struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
    FollowerID  uuid.UUID `gorm:"type:uuid;not null;index"`
    FollowingID uuid.UUID `gorm:"type:uuid;not null;index"`
    CreatedAt   time.Time
}

type User struct{

	ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name 		string `gorm:"not null"`
	Username 	string `gorm:"unique;not null"`
	Email 		string `gorm:"unique;not null"`
	Password 	string `gorm:"not null"`
	Bio string

	DateOfBirth     time.Time
	Gender          string
	ProfilePictureURL string
	SubscribedToNewsletter bool `gorm:"default:false"`
	TwoFactorEnabled bool `gorm:"default:false"`
	IsActive bool `gorm:"default:true"` 
	IsBanned bool `gorm:"default:false"`

	Followers []Follow `gorm:"foreignKey:FollowingID"`
    Following []Follow `gorm:"foreignKey:FollowerID"`
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	// Generate a new UUID (version 4)
	user.ID = uuid.New()

	if user.Username == ""{
		user.Username = user.ID.String()
	}
	return
}

func (follow *Follow) BeforeCreate(tx *gorm.DB) (err error) {
	follow.ID = uuid.New()
	return
}