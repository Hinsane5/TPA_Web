package domain

import (
	"time"
	"github.com/google/uuid"
)

type Post struct {
	ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	MediaObjectName string    `gorm:"type:varchar(255);not null"` 
	MediaType       string    `gorm:"type:varchar(50)"`       
	Caption         string    `gorm:"type:text"`
	Location        string    `gorm:"type:varchar(255)"`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}