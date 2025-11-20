package domain

import (
	"time"
	"github.com/google/uuid"
)

type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    // For group chats
	IsGroup   bool      `gorm:"default:false"`
	CreatedAt time.Time
	
	// Relations
	Participants []Participant `gorm:"foreignKey:ConversationID"`
	Messages     []Message     `gorm:"foreignKey:ConversationID"`
}

type Participant struct {
	ConversationID uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	JoinedAt       time.Time
	LastReadMessageID *uuid.UUID `gorm:"type:uuid"`
}

type Message struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null"`
	Content        string    
	MediaURL       string
	MediaType      string    // 'text', 'image', 'video'
	IsUnsent       bool      `gorm:"default:false"`
	CreatedAt      time.Time
}