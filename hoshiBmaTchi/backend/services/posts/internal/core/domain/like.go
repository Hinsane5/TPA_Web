package domain

import (
	"time"

	"github.com/google/uuid"
)

type PostLike struct {
	UserID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	PostID  uuid.UUID `gorm:"type:uuid;primaryKey"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (PostLike) TableName() string {
	return "post_likes"
}