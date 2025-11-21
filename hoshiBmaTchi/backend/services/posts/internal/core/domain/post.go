package domain

import (
	"time"
	"github.com/google/uuid"
)

type PostMedia struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	PostID          uuid.UUID `gorm:"type:uuid;not null"`
	MediaObjectName string    `gorm:"type:varchar(255);not null"`
	MediaType       string    `gorm:"type:varchar(50)"`
	Sequence        int       `gorm:"type:int"` 
}

type Post struct {
	ID              uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID   `gorm:"type:uuid;not null"`
	
    // --- THIS FIELD WAS MISSING ---
	Media           []PostMedia `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE"` 
    // ------------------------------

	Caption         string      `gorm:"type:text"`
	Location        string      `gorm:"type:varchar(255)"`
	CreatedAt       time.Time   `gorm:"autoCreateTime"`
	UpdatedAt       time.Time   `gorm:"autoUpdateTime"`
	LikesCount      int32       `gorm:"-"`
	CommentsCount   int32       `gorm:"-"`
	IsLiked         bool        `gorm:"-"`
}

type SavedPost struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null"`
	PostID       uuid.UUID `gorm:"type:uuid;not null"`
	CollectionID uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	
	Post Post `gorm:"foreignKey:PostID"`
}

type Collection struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"type:varchar(100);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	
	// Relationships
	SavedPosts []SavedPost `gorm:"foreignKey:CollectionID"`
}