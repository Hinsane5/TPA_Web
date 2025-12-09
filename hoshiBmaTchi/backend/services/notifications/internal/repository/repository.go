package repository

import (
	"github.com/Hinsane5/hoshiBmaTchi/backend/services/notifications/internal/models"
	"gorm.io/gorm"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *NotificationRepository) GetByUserID(userID uint) ([]models.Notification, error) {
	var notifications []models.Notification
	// Fetch latest 50, ordered by time
	err := r.db.Where("recipient_id = ?", userID).Order("created_at desc").Limit(50).Find(&notifications).Error
	return notifications, err
}