package domain

import (
	"time"
	"github.com/google/uuid"
)

type Conversation struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Name      string    `json:"name"`
	IsGroup   bool      `gorm:"default:false" json:"is_group"`
	CreatedAt time.Time `json:"created_at"`
	
	Participants []Participant `gorm:"foreignKey:ConversationID" json:"participants"`
	Messages     []Message     `gorm:"foreignKey:ConversationID" json:"messages"`
}

type Participant struct {
	ConversationID    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"conversation_id"`
	UserID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"` 
	JoinedAt          time.Time  `json:"joined_at"`
	LastReadMessageID *uuid.UUID `gorm:"type:uuid" json:"last_read_message_id"`
}

type Message struct {
	ID             uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null" json:"conversation_id"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content        string    `json:"content"`
	MediaURL       string    `json:"media_url"`
	MediaType      string    `json:"media_type"` 
	IsUnsent       bool      `gorm:"default:false" json:"is_unsent"`
	CreatedAt      time.Time `json:"created_at"`
}