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

type Block struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
    BlockerID uuid.UUID `gorm:"type:uuid;not null;index"`
    BlockedID uuid.UUID `gorm:"type:uuid;not null;index"`
    CreatedAt time.Time
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

	PushNotificationsEnabled bool `gorm:"default:true"`
    EmailNotificationsEnabled bool `gorm:"default:true"`

	IsPrivate       bool `gorm:"default:false"`

	CloseFriends []CloseFriend `gorm:"foreignKey:UserID"`
    HiddenStoryViewers []HiddenStoryViewer `gorm:"foreignKey:UserID"`
    VerificationRequests []VerificationRequest `gorm:"foreignKey:UserID"`

	Role       string `gorm:"default:'user'"`
    IsVerified bool   `gorm:"default:false"`
}

type CloseFriend struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
    UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
    CloseFriendID uuid.UUID `gorm:"type:uuid;not null;index"`
    CreatedAt   time.Time
}

type HiddenStoryViewer struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;"`
    UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
    HiddenUserID uuid.UUID `gorm:"type:uuid;not null;index"`
    CreatedAt   time.Time
}

type VerificationRequest struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;"`
    UserID          uuid.UUID `gorm:"type:uuid;not null"`
    NationalIDNumber string   `gorm:"not null"`
    Reason          string    `gorm:"type:text"`
    SelfieURL       string
    Status          string    `gorm:"default:'PENDING'"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type UserReport struct {
    ID             uuid.UUID `gorm:"type:uuid;primary_key;"`
    ReporterID     uuid.UUID `gorm:"type:uuid;not null"`
    ReportedUserID uuid.UUID `gorm:"type:uuid;not null"`
    Reason         string
    Status         string    `gorm:"default:'PENDING'"`
    CreatedAt      time.Time
}

type PostReport struct {
    ID             uuid.UUID `gorm:"type:uuid;primary_key;"`
    ReporterID     uuid.UUID `gorm:"type:uuid;not null"`
    PostID         uuid.UUID `gorm:"type:uuid;not null"`
    Reason         string
    Status         string    `gorm:"default:'PENDING'"`
    CreatedAt      time.Time
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
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

func (block *Block) BeforeCreate(tx *gorm.DB) (err error) {
    block.ID = uuid.New()
    return
}