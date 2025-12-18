package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MediaType string

const (
	MediaTypeImage MediaType = "IMAGE"
	MediaTypeVideo MediaType = "VIDEO"
)


type Story struct {
	ID         string    `gorm:"type:uuid;primary_key"`
	UserID     string    `gorm:"type:uuid;not null;index"`
	MediaURL   string    `gorm:"type:text;not null"`
	MediaType  MediaType `gorm:"type:varchar(10);not null"`
	Duration   int       `gorm:"default:5"`
	Views []StoryView `gorm:"foreignKey:StoryID;constraint:OnDelete:CASCADE;"`
    Likes []StoryLike `gorm:"foreignKey:StoryID;constraint:OnDelete:CASCADE;"`
	ReplyCount int       `gorm:"default:0"`
	CreatedAt  time.Time `gorm:"index"`
	ExpiresAt  time.Time `gorm:"index"`
	DeletedAt  gorm.DeletedAt
}

func (s *Story) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	if s.ExpiresAt.IsZero() {
		s.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	return nil
}

type StoryView struct {
	ID        string    `gorm:"type:uuid;primary_key"`
	StoryID   string    `gorm:"type:uuid;not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	ViewedAt  time.Time `gorm:"not null"`
	DeletedAt gorm.DeletedAt
}

func (sv *StoryView) BeforeCreate(tx *gorm.DB) error {
	if sv.ID == "" {
		sv.ID = uuid.New().String()
	}
	if sv.ViewedAt.IsZero() {
		sv.ViewedAt = time.Now()
	}
	return nil
}

type StoryLike struct {
	ID        string    `gorm:"type:uuid;primary_key"`
	StoryID   string    `gorm:"type:uuid;not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	LikedAt   time.Time `gorm:"autoCreateTime"`
}

func (sl *StoryLike) BeforeCreate(tx *gorm.DB) error {
	if sl.ID == "" {
		sl.ID = uuid.New().String()
	}
	if sl.LikedAt.IsZero() {
		sl.LikedAt = time.Now()
	}
	return nil
}

type StoryReply struct {
	ID        string    `gorm:"type:uuid;primary_key"`
	StoryID   string    `gorm:"type:uuid;not null;index"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Content   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	DeletedAt gorm.DeletedAt
}

func (sr *StoryReply) BeforeCreate(tx *gorm.DB) error {
	if sr.ID == "" {
		sr.ID = uuid.New().String()
	}
	if sr.CreatedAt.IsZero() {
		sr.CreatedAt = time.Now()
	}
	return nil
}

type StoryShare struct {
	ID          string    `gorm:"type:uuid;primary_key"`
	StoryID     string    `gorm:"type:uuid;not null;index"`
	SenderID    string    `gorm:"type:uuid;not null;index"`
	RecipientID string    `gorm:"type:uuid;not null;index"`
	MessageID   string    `gorm:"type:uuid"`
	CreatedAt   time.Time `gorm:"not null"`
	DeletedAt   gorm.DeletedAt
}

func (ss *StoryShare) BeforeCreate(tx *gorm.DB) error {
	if ss.ID == "" {
		ss.ID = uuid.New().String()
	}
	if ss.CreatedAt.IsZero() {
		ss.CreatedAt = time.Now()
	}
	return nil
}

type StoryVisibility struct {
    ID             string `gorm:"type:uuid;primary_key"`
    UserID         string `gorm:"type:uuid;index"`
    HiddenViewerID string `gorm:"type:uuid;index"`
}

