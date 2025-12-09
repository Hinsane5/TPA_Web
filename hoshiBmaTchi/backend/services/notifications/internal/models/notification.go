package models

import "gorm.io/gorm"

type Notification struct {
	gorm.Model
	RecipientID string      `json:"recipient_id" gorm:"index"` 
	SenderID    string      `json:"sender_id"`                
	SenderName  string    `json:"sender_name"`               
	SenderImage string    `json:"sender_image"`              
	Type        string    `json:"type"`                      
	EntityID    string      `json:"entity_id"`                 
	Message     string    `json:"message"`                   
	IsRead      bool      `json:"is_read" gorm:"default:false"`
}

type NotificationEvent struct {
	RecipientID string   `json:"recipient_id"`
	SenderID    string   `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	SenderImage string `json:"sender_image"`
	Type        string `json:"type"`      
	EntityID    string   `json:"entity_id"` 
	Message     string `json:"message"`
}